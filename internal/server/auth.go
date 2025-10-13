package server

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// --- Constants for configuration and clarity ---
const (
	// EVE API URLs
	eveVerifyURL = "https://login.eveonline.com/oauth/verify"
	esiCharURL   = "https://esi.evetech.net/latest/characters/%d/"

	// Application-specific settings
	sessionName        = "wingspan-session"
	sessionStateKey    = "oauth_state"
	sessionAuthKey     = "authenticated"
	sessionCharNameKey = "character_name"

	// Wingspan Corporation ID
	wingspanCorpID = 1000182
)

// --- Structs for decoding EVE API responses ---

type EveVerifyResponse struct {
	CharacterID   int    `json:"CharacterID"`
	CharacterName string `json:"CharacterName"`
}

type EsiCharacterResponse struct {
	CorporationID int `json:"corporation_id"`
}

// --- HTTP Handlers ---

// loginHandler starts the EVE SSO process.
func (s *Server) loginHandler(w http.ResponseWriter, r *http.Request) {
	session, err := s.sessionStore.Get(r, sessionName)
	if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}

	// Generate a secure random state token
	state, err := generateRandomState()
	if err != nil {
		log.Printf("ERROR: Failed to generate state token: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	session.Values[sessionStateKey] = state
	if err := session.Save(r, w); err != nil {
		log.Printf("ERROR: Failed to save session: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	redirectURL := s.oauthConfig.AuthCodeURL(state)
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

// callbackHandler is the final step in the SSO flow.
// It has been refactored to be a coordinator, calling helper functions for each step.
func (s *Server) callbackHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Validate the state token to prevent CSRF attacks.
	if err := s.validateState(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 2. Exchange the authorization code for a token and verify the character.
	verifyResponse, err := s.verifyEveSSO(r.Context(), r.FormValue("code"))
	if err != nil {
		log.Printf("ERROR: EVE SSO verification failed: %v", err)
		http.Error(w, "Failed to verify EVE character", http.StatusInternalServerError)
		return
	}

	// 3. Check if the character is a member of the required corporation.
	isMember, err := s.isWingspanMember(verifyResponse.CharacterID)
	if err != nil {
		log.Printf("ERROR: Corporation check failed for char ID %d: %v", verifyResponse.CharacterID, err)
		http.Error(w, "Failed to check character's corporation", http.StatusInternalServerError)
		return
	}
	if !isMember {
		log.Printf("ACCESS DENIED: %s (ID: %d) is not in WINGSPAN.", verifyResponse.CharacterName, verifyResponse.CharacterID)
		http.Error(w, "Access Denied: This platform is for Wingspan members only.", http.StatusForbidden)
		return
	}

	// 4. All checks passed. Log the user in by updating the session.
	session, _ := s.sessionStore.Get(r, sessionName) // We can ignore this error as it was checked in validateState.
	session.Values[sessionAuthKey] = true
	session.Values[sessionCharNameKey] = verifyResponse.CharacterName
	if err := session.Save(r, w); err != nil {
		log.Printf("ERROR: Failed to save final session: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("User logged in: %s (ID: %d)", verifyResponse.CharacterName, verifyResponse.CharacterID)
	http.Redirect(w, r, "/", http.StatusFound)
}

// logoutHandler clears the session and logs the user out.
func (s *Server) logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, err := s.sessionStore.Get(r, sessionName)
	if err != nil {
		// Even if getting the session fails, we can still try to redirect.
		http.Redirect(w, r, "/static/login.html", http.StatusFound)
		return
	}

	session.Values[sessionAuthKey] = false
	session.Options.MaxAge = -1 // This effectively deletes the cookie.
	session.Save(r, w)

	http.Redirect(w, r, "/static/login.html", http.StatusFound)
}

// authMiddleware protects routes that require a valid login session.
func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := s.sessionStore.Get(r, sessionName)
		if err != nil {
			// If we can't get a session, they are not authenticated.
			http.Redirect(w, r, "/static/login.html", http.StatusFound)
			return
		}

		if auth, ok := session.Values[sessionAuthKey].(bool); !ok || !auth {
			http.Redirect(w, r, "/static/login.html", http.StatusFound)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// --- Helper Functions ---

// validateState checks the state token from the callback against the one in the session.
func (s *Server) validateState(r *http.Request) error {
	session, err := s.sessionStore.Get(r, sessionName)
	if err != nil {
		return fmt.Errorf("failed to get session")
	}

	originalState, ok := session.Values[sessionStateKey].(string)
	if !ok || originalState == "" || r.FormValue("state") != originalState {
		return fmt.Errorf("invalid state token")
	}
	return nil
}

// verifyEveSSO handles the OAuth token exchange and fetches the character's identity.
func (s *Server) verifyEveSSO(ctx context.Context, code string) (*EveVerifyResponse, error) {
	token, err := s.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange token: %w", err)
	}

	client := s.oauthConfig.Client(ctx, token)
	resp, err := client.Get(eveVerifyURL)
	if err != nil {
		return nil, fmt.Errorf("failed to call verify endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("verify endpoint returned non-200 status: %s", resp.Status)
	}

	var verifyResponse EveVerifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&verifyResponse); err != nil {
		return nil, fmt.Errorf("failed to decode verify response: %w", err)
	}

	return &verifyResponse, nil
}

// isWingspanMember checks if a character is part of the designated corporation.
func (s *Server) isWingspanMember(characterID int) (bool, error) {
	url := fmt.Sprintf(esiCharURL, characterID)
	resp, err := http.Get(url)
	if err != nil {
		return false, fmt.Errorf("failed to call ESI character endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("ESI endpoint returned non-200 status: %s", resp.Status)
	}

	var charResponse EsiCharacterResponse
	if err := json.NewDecoder(resp.Body).Decode(&charResponse); err != nil {
		return false, fmt.Errorf("failed to decode ESI character response: %w", err)
	}

	return charResponse.CorporationID == wingspanCorpID, nil
}

// generateRandomState creates a cryptographically secure random string for the state token.
func generateRandomState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

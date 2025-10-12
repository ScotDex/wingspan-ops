//go:build ignore

package server

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
)

// These structs are for decoding the responses from EVE's SSO.
type EveVerifyResponse struct {
	CharacterID   int    `json:"CharacterID"`
	CharacterName string `json:"CharacterName"`
}
type EsiCharacterResponse struct {
	CorporationID int `json:"corporation_id"`
}

// loginHandler starts the EVE SSO process. It's now a method on the Server.
func (s *Server) loginHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := s.sessionStore.Get(r, "wingspan-session")

	// Generate a random state token for security
	b := make([]byte, 32)
	rand.Read(b)
	state := base64.StdEncoding.EncodeToString(b)

	session.Values["oauth_state"] = state
	session.Save(r, w)

	redirectURL := s.oauthConfig.AuthCodeURL(state)
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

// callbackHandler handles the redirect back from EVE's SSO.
func (s *Server) callbackHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := s.sessionStore.Get(r, "wingspan-session")
	originalState, ok := session.Values["oauth_state"].(string)
	if !ok || originalState == "" || r.FormValue("state") != originalState {
		http.Error(w, "Invalid state token", http.StatusBadRequest)
		return
	}

	token, err := s.oauthConfig.Exchange(r.Context(), r.FormValue("code"))
	if err != nil {
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		return
	}

	client := s.oauthConfig.Client(r.Context(), token)
	resp, err := client.Get("https://login.eveonline.com/oauth/verify")
	if err != nil {
		http.Error(w, "Failed to verify character", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var verifyResponse EveVerifyResponse
	json.NewDecoder(resp.Body).Decode(&verifyResponse)

	// Here you would add your logic to check the character's corporation ID.
	// For now, we'll just log them in.
	log.Printf("User logged in: %s (ID: %d)", verifyResponse.CharacterName, verifyResponse.CharacterID)

	session.Values["authenticated"] = true
	session.Values["character_name"] = verifyResponse.CharacterName
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusFound) // Redirect to the main page
}

// logoutHandler clears the session.
func (s *Server) logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := s.sessionStore.Get(r, "wingspan-session")
	session.Values["authenticated"] = false
	session.Options.MaxAge = -1 // Deletes the cookie
	session.Save(r, w)
	http.Redirect(w, r, "/static/login.html", http.StatusFound)
}

// authMiddleware protects routes that require a login.
func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := s.sessionStore.Get(r, "wingspan-session")
		if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
			http.Redirect(w, r, "/static/login.html", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}

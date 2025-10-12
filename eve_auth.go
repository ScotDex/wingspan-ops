//go:build ignore

import (
    // ... other imports
    "github.com/gorilla/sessions"
	"golang.org/x/oauth2"
)

type Server struct {
	// ... (templates, urls, esiClient, graph)
	oauthConfig  *oauth2.Config        // ADD THIS
	sessionStore *sessions.CookieStore // ADD THIS
}

func New(..., graph *routing.Graph, oauthConfig *oauth2.Config, sessionStore *sessions.CookieStore) (*Server, error) { // MODIFY
	// ...
	return &Server{
		// ... (other fields)
		graph:        graph,
		oauthConfig:  oauthConfig,  // ADD THIS
		sessionStore: sessionStore, // ADD THIS
	}, nil
}


import (
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
    // ...
)

func main() {
    // ... (godotenv.Load(), other config)

	// --- ADD THIS BLOCK ---
	// Add these secrets to your .env file
	eveClientID := os.Getenv("EVE_CLIENT_ID")
	eveClientSecret := os.Getenv("EVE_CLIENT_SECRET")
	eveCallbackURL := os.Getenv("EVE_CALLBACK_URL") // e.g., http://localhost:8080/auth/callback
	sessionSecretKey := os.Getenv("SESSION_SECRET_KEY")

	oauthConfig := &oauth2.Config{
		RedirectURL:  eveCallbackURL,
		ClientID:     eveClientID,
		ClientSecret: eveClientSecret,
		Scopes:       []string{}, // Add any ESI scopes you need here
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://login.eveonline.com/v2/oauth/authorize",
			TokenURL: "https://login.eveonline.com/v2/oauth/token",
		},
	}
	sessionStore := sessions.NewCookieStore([]byte(sessionSecretKey))
	// ----------------------

	// Pass the new objects to your server
	srv, err := server.New(wingspanURL, feedbackURL, donateURL, esiClient, graph, oauthConfig, sessionStore)
    // ...
}
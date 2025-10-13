package main

import (
	"log"
	"net/http"
	"os"
	"sync"
	"wingspan-ops/internal/esi"
	"wingspan-ops/internal/routing"
	"wingspan-ops/internal/server"
	"wingspan-ops/internal/updater"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

func main() {
	// Load environment variables from a .env file if it exists.
	err := godotenv.Load()
	if err != nil {
		log.Println("Note: No .env file found, reading from OS environment.")
	}

	// --- Authentication Setup ---
	// 1. Initialize the session store with a secret key from the environment.
	sessionKey := os.Getenv("SESSION_KEY")
	if sessionKey == "" {
		log.Fatal("FATAL: SESSION_KEY environment variable not set. Please provide a 32 or 64 byte key.")
	}
	sessionStore := sessions.NewCookieStore([]byte(sessionKey))

	// 2. Initialize the OAuth2 config with your EVE application credentials.
	clientID := os.Getenv("EVE_CLIENT_ID")
	secretKey := os.Getenv("EVE_SECRET_KEY")
	if clientID == "" || secretKey == "" {
		log.Fatal("FATAL: EVE_CLIENT_ID and EVE_SECRET_KEY environment variables must be set.")
	}
	oauthConfig := &oauth2.Config{
		RedirectURL:  os.Getenv("EVE_CALLBACK_URL"), // e.g., http://localhost:8080/callback
		ClientID:     clientID,
		ClientSecret: secretKey,
		Scopes:       []string{}, // No specific scopes needed for just identity
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://login.eveonline.com/v2/oauth/authorize",
			TokenURL: "https://login.eveonline.com/v2/oauth/token",
		},
	}
	if oauthConfig.RedirectURL == "" {
		log.Fatal("FATAL: EVE_CALLBACK_URL environment variable not set.")
	}
	// --- End Authentication Setup ---

	// Initialize the ESI client for fetching game data.
	esiClient := esi.NewESIClient("themadlyscientific@gmail.com")
	if err := esiClient.LoadSystemNameCache("systems.json"); err != nil {
		log.Printf("WARN: Could not load local system name cache: %v", err)
	}

	// Load required application configuration from environment variables.
	feedbackURL := os.Getenv("FEEDBACK_FORM_URL")
	if feedbackURL == "" {
		log.Fatal("FATAL: FEEDBACK_FORM_URL environment variable not set.")
	}
	wingspanURL := os.Getenv("WINGSPAN_API_URL")
	if wingspanURL == "" {
		log.Fatal("FATAL: WINGSPAN_API_URL environment variable not set.")
	}

	// Load the static stargate map data for routing.
	graph := routing.NewGraph()
	if err := graph.LoadCSV("mapSolarSystemJumps.csv"); err != nil {
		log.Fatalf("FATAL: Could not load stargate map: %v", err)
	}
	log.Printf("✅ Loaded %d systems into the static stargate graph.", graph.StaticAdjacencyListSize())

	// Start a background process to update EVE Online kill data.
	var wg sync.WaitGroup
	killUpdater := updater.New(esiClient, "kills.json")
	wg.Add(1)
	go killUpdater.Start(&wg)

	// Create the main server instance, now with auth components.
	srv, err := server.New(
		wingspanURL,
		feedbackURL,
		esiClient,
		graph,
		oauthConfig,
		sessionStore,
	)
	if err != nil {
		log.Fatalf("FATAL: Failed to create server: %v", err)
	}

	// Register all the HTTP routes.
	router := srv.RegisterRoutes()

	// Determine the port and start the web server.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("🚀 Starting server on http://localhost:%s", port)
	if err = http.ListenAndServe(":"+port, router); err != nil {
		log.Fatalf("FATAL: Failed to start server: %v", err)
	}
}

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

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Note: No .env file found, reading from OS environment.")
	}

	esiClient := esi.NewESIClient("themadlyscientific@gmail.com")

	if err := esiClient.LoadSystemNameCache("systems.json"); err != nil {
		log.Printf("WARN: Could not load local system name cache: %v", err)
	}

	feedbackURL := os.Getenv("FEEDBACK_FORM_URL")
	if feedbackURL == "" {
		log.Fatal("FATAL: FEEDBACK_FORM_URL environment variable not set.")
	}

	wingspanURL := os.Getenv("WINGSPAN_API_URL")
	if wingspanURL == "" {
		log.Fatal("FATAL: WINGSPAN_API_URL environment variable not set.")
	}

	graph := routing.NewGraph()
	if err := graph.LoadCSV("mapSolarSystemJumps.csv"); err != nil {
		log.Fatalf("FATAL: Could not load stargate map: %v", err)
	}

	var wg sync.WaitGroup
	killUpdater := updater.New(esiClient, "kills.json")
	wg.Add(1)
	go killUpdater.Start(&wg)

	// UPDATED: Call the new getter method.
	log.Printf("✅ Loaded %d systems into the static stargate graph.", graph.StaticAdjacencyListSize())

	// UPDATED: Pass the 'graph' object to the server constructor.
	srv, err := server.New(wingspanURL, feedbackURL, esiClient, graph)
	if err != nil {
		log.Fatalf("FATAL: Failed to create server: %v", err)
	}

	router := srv.RegisterRoutes()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Starting server on http://localhost:%s", port)
	if err = http.ListenAndServe(":"+port, router); err != nil {
		log.Fatalf("FATAL: Failed to start server: %v", err)
	}
}

package updater // Change the package name

import (
	"context" // Add context for the ESI call
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"
	"wingspan-ops/internal/esi" // Import your esi package
)

// Renamed to Updater for simplicity
type Updater struct {
	esiClient *esi.ESIClient
	filePath  string
}

// Renamed to New
func New(client *esi.ESIClient, filePath string) *Updater {
	return &Updater{
		esiClient: client,
		filePath:  filePath,
	}
}

func (u *Updater) Start(wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("[UPDATER] Starting background kill data updater...")
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	u.fetchAndSave() // Run once on startup

	for {
		<-ticker.C
		u.fetchAndSave()
	}
}

func (u *Updater) fetchAndSave() {
	log.Println("[UPDATER] Fetching latest system kill data from ESI...")
	// Pass a context to the ESI call
	kills, err := u.esiClient.GetSystemKills(context.Background())
	if err != nil {
		log.Printf("[UPDATER] ERROR: Failed to fetch kills from ESI: %v", err)
		return
	}

	jsonData, err := json.Marshal(kills)
	if err != nil {
		log.Printf("[UPDATER] ERROR: Failed to marshal kills to JSON: %v", err)
		return
	}

	tempFilePath := u.filePath + ".tmp"
	if err := os.WriteFile(tempFilePath, jsonData, 0644); err != nil {
		log.Printf("[UPDATER] ERROR: Failed to write to temp file: %v", err)
		return
	}

	if err := os.Rename(tempFilePath, u.filePath); err != nil {
		log.Printf("[UPDATER] ERROR: Failed to rename temp file: %v", err)
		return
	}

	log.Printf("[UPDATER] ✅ Successfully saved kill data to %s.", u.filePath)
}

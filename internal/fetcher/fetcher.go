package fetcher

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	"wingspan-ops/internal/models"
	// Use your module name
)

// CORRECTED: This function now returns the entire API response struct.
func FetchWingspanData(apiURL string) (*models.WingspanAPIResponse, error) {
	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to make request to wingspan api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("wingspan api returned non-200 status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var apiResponse models.WingspanAPIResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal json: %w", err)
	}

	// Return a pointer to the whole struct.
	return &apiResponse, nil
}

func FetchTheraData() ([]models.TheraConnection, error) {
	// Use a custom client with a timeout for robustness.
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://api.eve-scout.com/v2/public/signatures?system_name=thera")
	if err != nil {
		return nil, fmt.Errorf("failed to make request to eve-scout api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("eve-scout api returned non-200 status: %d", resp.StatusCode)
	}

	var theraConnections []models.TheraConnection
	if err := json.NewDecoder(resp.Body).Decode(&theraConnections); err != nil {
		return nil, fmt.Errorf("failed to unmarshal eve-scout json: %w", err)
	}

	return theraConnections, nil
}

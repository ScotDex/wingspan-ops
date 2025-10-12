package server

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
	"wingspan-ops/internal/esi"
	"wingspan-ops/internal/models"
)

func formatSignatureID(sigID string) string {
	// Only format if the string is exactly 6 characters long and has no dash.
	if len(sigID) == 6 && !strings.Contains(sigID, "-") {
		return fmt.Sprintf("%s-%s", sigID[:3], sigID[3:])
	}
	// Otherwise, return the original string.
	return sigID
}

func processAPIResponse(response *models.WingspanAPIResponse, esiClient *esi.ESIClient, killMap map[int]esi.EsiSystemKills) ([]models.ConnectionInfo, []models.LeaderboardEntry) {
	var connections []models.ConnectionInfo
	scanCounts := make(map[string]int)

	for _, wh := range response.Wormholes {
		sigInitial, ok1 := response.Signatures[wh.InitialID]
		sigSecondary, ok2 := response.Signatures[wh.SecondaryID]

		if ok1 && ok2 {
			scanCounts[sigInitial.CreatedByName]++

			// Convert string IDs to integers.
			fromID, _ := strconv.Atoi(sigInitial.SystemID)
			toID, _ := strconv.Atoi(sigSecondary.SystemID)

			// Use the cached GetSystemName function directly.
			// This is fast because it checks your local file first.
			fromName := esiClient.GetSystemName(fromID)
			toName := esiClient.GetSystemName(toID)
			var signatureID string
			if sigInitial.SignatureID != nil {
				signatureID = strings.ToUpper(*sigInitial.SignatureID)
			}
			if fromName != "Unknown" && toName != "Unknown" && signatureID != "???" {
				connection := models.ConnectionInfo{
					FromName:    fromName,
					ToName:      toName,
					Scout:       sigInitial.CreatedByName,
					LastUpdated: sigInitial.ModifiedTime,
					Eol:         wh.Life,
					EolStatus:   wh.Life,
					SignatureID: formatSignatureID(signatureID),
				}

				connections = append(connections, connection)
			}
		}
	}

	// Leaderboard logic (unchanged)
	var leaderboard []models.LeaderboardEntry
	for name, count := range scanCounts {
		leaderboard = append(leaderboard, models.LeaderboardEntry{ScoutName: name, ScanCount: count})
	}
	sort.Slice(leaderboard, func(i, j int) bool {
		return leaderboard[i].ScanCount > leaderboard[j].ScanCount
	})

	return connections, leaderboard
}

// processTheraConnections remains the same
func processTheraConnections(theraConns []models.TheraConnection) []models.ConnectionInfo {
	var connections []models.ConnectionInfo
	for _, tc := range theraConns {
		connection := models.ConnectionInfo{
			FromName:    tc.OutSystemName,
			ToName:      tc.InSystemName,
			SignatureID: tc.InSignature,
			Eol:         fmt.Sprintf("%d hours", tc.RemainingHours),
			Scout:       tc.CreatedByName,
			LastUpdated: tc.CreatedAt.Format(time.RFC1123),
		}
		connections = append(connections, connection)
	}
	return connections
}

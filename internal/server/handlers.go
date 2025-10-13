package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"wingspan-ops/internal/esi"
	"wingspan-ops/internal/fetcher"
	"wingspan-ops/internal/models"
	"wingspan-ops/internal/routing"
)

// homeHandler is for your main "Live Map" page. It renders "index.html".
func (s *Server) homeHandler(w http.ResponseWriter, r *http.Request) {
	var allConnections []models.ConnectionInfo
	var leaderboard []models.LeaderboardEntry

	killMap := make(map[int]esi.EsiSystemKills)
	killData, err := os.ReadFile("kills.json")
	if err != nil {
		log.Printf("WARN: Could not read kills.json file: %v", err)
	} else {
		var kills []esi.EsiSystemKills
		if err := json.Unmarshal(killData, &kills); err == nil {
			for _, k := range kills {
				killMap[k.SystemID] = k
			}
		}
	}

	wingspanResponse, err := fetcher.FetchWingspanData(s.wingspanURL)
	if err != nil {
		log.Printf("WARN: Failed to fetch from Wingspan API: %v", err)
	} else {
		var wingspanConnections []models.ConnectionInfo
		wingspanConnections, leaderboard = processAPIResponse(wingspanResponse, s.esiClient, killMap)
		allConnections = append(allConnections, wingspanConnections...)
	}

	theraResponse, err := fetcher.FetchTheraData()
	if err != nil {
		log.Printf("WARN: Failed to fetch from Thera API: %v", err)
	} else {
		theraConnections := processTheraConnections(theraResponse)
		allConnections = append(allConnections, theraConnections...)
	}

	data := models.FrontendData{
		Connections: allConnections,
		Leaderboard: leaderboard,
		FeedbackURL: s.feedbackURL,
	}

	ts, ok := s.templates["index.html"]
	if !ok {
		http.Error(w, "Could not load index.html template", http.StatusInternalServerError)
		return
	}
	err = ts.Execute(w, data)
	if err != nil {
		log.Printf("Template execution error: %v", err)
	}
}

// shortCircuitHandler handles the route planning page and form submissions.
// shortCircuitHandler handles the route planning page and form submissions.
func (s *Server) shortCircuitHandler(w http.ResponseWriter, r *http.Request) {
	// Define the data struct at the top to be used by both GET and POST.
	data := models.FrontendData{
		FeedbackURL: s.feedbackURL,
	}

	if r.Method == http.MethodGet {
		ts, ok := s.templates["short_circuit.html"]
		if !ok {
			http.Error(w, "Could not load template", http.StatusInternalServerError)
			return
		}
		ts.Execute(w, data)
		return
	}

	if r.Method == http.MethodPost {
		startSystemName := r.FormValue("start_system")
		endSystemName := r.FormValue("end_system")

		startID, err := s.esiClient.GetSystemID(context.Background(), startSystemName)
		if err != nil {
			http.Error(w, fmt.Sprintf("Could not find start system: %s", startSystemName), http.StatusBadRequest)
			return
		}
		endID, err := s.esiClient.GetSystemID(context.Background(), endSystemName)
		if err != nil {
			http.Error(w, fmt.Sprintf("Could not find end system: %s", endSystemName), http.StatusBadRequest)
			return
		}

		killMap := make(map[int]esi.EsiSystemKills)
		killData, err := os.ReadFile("kills.json")
		if err != nil {
			log.Printf("WARN: Could not read kills.json file for routing: %v", err)
		} else {
			var kills []esi.EsiSystemKills
			if err := json.Unmarshal(killData, &kills); err == nil {
				for _, k := range kills {
					killMap[k.SystemID] = k
				}
			}
		}

		wingspanResponse, _ := fetcher.FetchWingspanData(s.wingspanURL)
		theraResponse, _ := fetcher.FetchTheraData()
		whLinks := processConnectionsToWHLinks(wingspanResponse, theraResponse, s.esiClient)

		requestGraph := s.graph.Clone()
		requestGraph.UpdateWormholes(whLinks)

		_, prev := requestGraph.ShortestPath(startID, endID)

		// Pass the esiClient and killMap to the helper function.
		path := reconstructPath(prev, startID, endID, s.esiClient, killMap)

		data.Path = path
		ts, ok := s.templates["short_circuit.html"]
		if !ok {
			http.Error(w, "Could not load template", http.StatusInternalServerError)
			return
		}
		ts.Execute(w, data)
	}
}

// lookupHandler handles the character lookup page and form submissions.
func (s *Server) lookupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		data := models.FrontendData{
			FeedbackURL: s.feedbackURL,
		}
		ts, ok := s.templates["lookup.html"]
		if !ok {
			http.Error(w, "Could not load lookup.html template", http.StatusInternalServerError)
			return
		}
		err := ts.Execute(w, data)
		if err != nil {
			log.Printf("Template execution error: %v", err)
		}
		return
	}

	if r.Method == http.MethodPost {
		charName := r.FormValue("character_name")
		if charName == "" {
			http.Error(w, "Character name cannot be empty", http.StatusBadRequest)
			return
		}
		charID, err := s.esiClient.GetCharacterID(context.Background(), charName)
		if err != nil {
			log.Printf("Failed to find character ID for '%s': %v", charName, err)
			http.Error(w, "Character not found.", http.StatusNotFound)
			return
		}
		killboardURL := fmt.Sprintf("https://eve-kill.com/character/%d", charID)
		http.Redirect(w, r, killboardURL, http.StatusSeeOther)
	}
}

func processConnectionsToWHLinks(wingspanResponse *models.WingspanAPIResponse, theraResponse []models.TheraConnection, esiClient *esi.ESIClient) []routing.WHLink {
	var links []routing.WHLink

	if wingspanResponse != nil {
		for _, wh := range wingspanResponse.Wormholes {
			sigInitial, ok1 := wingspanResponse.Signatures[wh.InitialID]
			sigSecondary, ok2 := wingspanResponse.Signatures[wh.SecondaryID]

			// This is the combined guard clause.
			// It checks everything at once: signatures exist, IDs are valid numbers, and IDs are valid systems.
			if !ok1 || !ok2 {
				continue // Skip if either signature is missing
			}
			fromID, err1 := strconv.Atoi(sigInitial.SystemID)
			toID, err2 := strconv.Atoi(sigSecondary.SystemID)
			if err1 != nil || err2 != nil || fromID < 30000000 || toID < 30000000 {
				continue // Skip if IDs are not valid numbers or not valid systems
			}

			// If all checks pass, add the link.
			links = append(links, routing.WHLink{From: fromID, To: toID, Cost: 1})
		}
	}

	if theraResponse != nil {
		for _, tc := range theraResponse {
			// Apply the same robust check to the Thera data.
			if tc.OutSystemID >= 30000000 && tc.InSystemID >= 30000000 {
				links = append(links, routing.WHLink{From: tc.OutSystemID, To: tc.InSystemID, Cost: 1})
			}
		}
	}

	return links
}

// --- Replace your existing reconstructPath function with this version ---
func reconstructPath(prev map[int]int, start, end int, esiClient *esi.ESIClient, killMap map[int]esi.EsiSystemKills) []models.PathStep {
	var path []models.PathStep
	current := end

	if _, ok := prev[current]; !ok && current != start {
		return nil // No path found
	}

	for {
		// Get full system details to find security status.
		sysInfo, err := esiClient.GetSystemDetails(current)

		// Look up the kill data for the current system from the map.
		kills := killMap[current]

		var step models.PathStep
		if err != nil {
			// If details aren't found, fallback to just the name.
			step = models.PathStep{
				SystemName: esiClient.GetSystemName(current),
				ShipKills:  kills.ShipKills,
				NpcKills:   kills.NpcKills,
			}
		} else {
			// Determine the security class for CSS.
			var secClass string
			if sysInfo.SecurityStatus >= 0.5 {
				secClass = "high-sec"
			} else if sysInfo.SecurityStatus > 0.0 {
				secClass = "low-sec"
			} else {
				secClass = "null-sec"
			}
			step = models.PathStep{
				SystemName:     sysInfo.Name,
				SecurityStatus: sysInfo.SecurityStatus,
				SecurityClass:  secClass,
				ShipKills:      kills.ShipKills,
				NpcKills:       kills.NpcKills,
			}
		}
		path = append(path, step)

		if current == start {
			break
		}
		current = prev[current]
	}

	// Reverse the path to go from start -> end
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}

	return path
}

// Add this new handler to the file

func (s *Server) aboutHandler(w http.ResponseWriter, r *http.Request) {
	// We still create the data struct to pass the URLs to the layout
	data := models.FrontendData{
		FeedbackURL: s.feedbackURL,
	}

	ts, ok := s.templates["about.html"]
	if !ok {
		http.Error(w, "Could not load about.html template", http.StatusInternalServerError)
		return
	}
	ts.Execute(w, data)
}

// loginPageHandler serves the static login page.
// We use a handler instead of relying solely on the file server
// to have a clean "/login" URL for redirects.
func (s *Server) loginPageHandler(w http.ResponseWriter, r *http.Request) {
	// In a real-world app, you might check if the user is already logged in
	// and redirect them to the homepage if they are.

	ts, ok := s.templates["login.html"]
	if !ok {
		http.Error(w, "Could not load login.html template", http.StatusInternalServerError)
		return
	}

	// This handler doesn't need to pass any dynamic data to the template.
	err := ts.Execute(w, nil)
	if err != nil {
		log.Printf("Template execution error: %v", err)
	}
}

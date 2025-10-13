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

// getAuthenticatedUser retrieves the character name from the session.
func (s *Server) getAuthenticatedUser(r *http.Request) string {
	session, err := s.sessionStore.Get(r, sessionName)
	if err != nil {
		return "" // No session found
	}

	if name, ok := session.Values[sessionCharNameKey].(string); ok {
		return name
	}

	return ""
}

// homeHandler renders the main "Live Map" page.
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
		Connections:   allConnections,
		Leaderboard:   leaderboard,
		FeedbackURL:   s.feedbackURL,
		CharacterName: s.getAuthenticatedUser(r),
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
func (s *Server) shortCircuitHandler(w http.ResponseWriter, r *http.Request) {
	data := models.FrontendData{
		FeedbackURL:   s.feedbackURL,
		CharacterName: s.getAuthenticatedUser(r),
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
			FeedbackURL:   s.feedbackURL,
			CharacterName: s.getAuthenticatedUser(r),
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

// aboutHandler renders the about page.
func (s *Server) aboutHandler(w http.ResponseWriter, r *http.Request) {
	data := models.FrontendData{
		FeedbackURL:   s.feedbackURL,
		CharacterName: s.getAuthenticatedUser(r),
	}

	ts, ok := s.templates["about.html"]
	if !ok {
		http.Error(w, "Could not load about.html template", http.StatusInternalServerError)
		return
	}
	ts.Execute(w, data)
}

func (s *Server) loginPageHandler(w http.ResponseWriter, r *http.Request) {
	// This handler's ONLY job is to render the login template.
	// The authMiddleware will handle redirecting users who are already logged in
	// because the /login route itself should also be protected by middleware
	// that redirects authenticated users away.

	ts, ok := s.templates["login.html"]
	if !ok {
		http.Error(w, "Could not load login.html template", http.StatusInternalServerError)
		return
	}

	err := ts.Execute(w, nil)
	if err != nil {
		log.Printf("Template execution error: %v", err)
	}
}

// --- Processing functions ---

func processConnectionsToWHLinks(wingspanResponse *models.WingspanAPIResponse, theraResponse []models.TheraConnection, esiClient *esi.ESIClient) []routing.WHLink {
	var links []routing.WHLink

	if wingspanResponse != nil {
		for _, wh := range wingspanResponse.Wormholes {
			sigInitial, ok1 := wingspanResponse.Signatures[wh.InitialID]
			sigSecondary, ok2 := wingspanResponse.Signatures[wh.SecondaryID]

			if !ok1 || !ok2 {
				continue
			}
			fromID, err1 := strconv.Atoi(sigInitial.SystemID)
			toID, err2 := strconv.Atoi(sigSecondary.SystemID)
			if err1 != nil || err2 != nil || fromID < 30000000 || toID < 30000000 {
				continue
			}

			links = append(links, routing.WHLink{From: fromID, To: toID, Cost: 1})
		}
	}

	if theraResponse != nil {
		for _, tc := range theraResponse {
			if tc.OutSystemID >= 30000000 && tc.InSystemID >= 30000000 {
				links = append(links, routing.WHLink{From: tc.OutSystemID, To: tc.InSystemID, Cost: 1})
			}
		}
	}

	return links
}

func reconstructPath(prev map[int]int, start, end int, esiClient *esi.ESIClient, killMap map[int]esi.EsiSystemKills) []models.PathStep {
	var path []models.PathStep
	current := end

	if _, ok := prev[current]; !ok && current != start {
		return nil // No path found
	}

	for {
		sysInfo, err := esiClient.GetSystemDetails(current)
		kills := killMap[current]
		var step models.PathStep
		if err != nil {
			step = models.PathStep{
				SystemName: esiClient.GetSystemName(current),
				ShipKills:  kills.ShipKills,
				NpcKills:   kills.NpcKills,
			}
		} else {
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

	// Reverse the path
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}

	return path
}

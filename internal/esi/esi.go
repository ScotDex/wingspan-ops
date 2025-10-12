package esi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// --- Structs ---

type esiName struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
}

type esiSystemIDResult struct {
	Systems []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"solar_systems"`
}

type ESISystemInfo struct {
	Name           string  `json:"name"`
	SecurityStatus float64 `json:"security_status"`
	SystemID       int     `json:"system_id"`
}

type esiCharacterIDResult struct {
	Characters []struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	} `json:"characters"`
}

// ESIClient manages all communication with the EVE Online ESI.
type ESIClient struct {
	httpClient      *http.Client
	baseURL         string
	userAgent       string
	cacheMutex      sync.RWMutex
	systemNameCache map[string]string      // ID -> Name (from local file)
	systemIDCache   map[string]int         // Name -> ID (from local file)
	nameCache       map[int]string         // ID -> Name (from live API calls)
	systemInfoCache map[int]*ESISystemInfo // ID -> Full Info (from local file)
}

// --- Constructor ---

func NewESIClient(contactInfo string) *ESIClient {
	return &ESIClient{
		httpClient:      &http.Client{Timeout: 15 * time.Second},
		baseURL:         "https://esi.evetech.net",
		userAgent:       fmt.Sprintf("Wingspan-Short-Circuit/1.0 (%s)", contactInfo),
		nameCache:       make(map[int]string),
		systemNameCache: make(map[string]string),
		systemIDCache:   make(map[string]int),
		systemInfoCache: make(map[int]*ESISystemInfo),
	}
}

// --- CONSOLIDATED CACHE LOADING ---

// LoadAllCachesFromSDE loads all necessary system data from a single detailed JSON file.
func (c *ESIClient) LoadSystemNameCache(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open SDE cache file: %w", err)
	}
	defer file.Close()

	// The JSON is a map of STRING keys to ESISystemInfo objects
	var tempCache map[string]*ESISystemInfo
	if err := json.NewDecoder(file).Decode(&tempCache); err != nil {
		return fmt.Errorf("failed to decode SDE cache: %w", err)
	}

	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()

	// Loop once and populate all three caches from the single source file.
	for idStr, data := range tempCache {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			continue // Skip if key is not a valid integer
		}

		// 1. Populate ID -> Full Info cache
		c.systemInfoCache[id] = data

		// 2. Populate ID -> Name cache
		c.systemNameCache[idStr] = data.Name

		// 3. Populate Name -> ID cache
		c.systemIDCache[strings.ToLower(data.Name)] = id
	}

	log.Printf("✅ Loaded all %d systems into local caches.", len(c.systemInfoCache))
	return nil
}

// --- Core HTTP Helper ---
func (c *ESIClient) do(ctx context.Context, method, endpoint string, body io.Reader, target any) error {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+endpoint, body)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")
	if method == http.MethodPost {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ESI returned non-200 status: %s", resp.Status)
	}

	return json.NewDecoder(resp.Body).Decode(target)
}

// --- Public Methods ---

// GetSystemID resolves a name to its ID, using the local cache first.
func (c *ESIClient) GetSystemID(ctx context.Context, name string) (int, error) {
	lowerName := strings.ToLower(name)
	c.cacheMutex.RLock()
	if id, ok := c.systemIDCache[lowerName]; ok {
		c.cacheMutex.RUnlock()
		return id, nil
	}
	c.cacheMutex.RUnlock()

	log.Printf("DEBUG: System name '%s' not in cache, fetching from ESI.", name)
	var idData esiSystemIDResult
	body, _ := json.Marshal([]string{name})
	err := c.do(ctx, http.MethodPost, "/universe/ids/", bytes.NewBuffer(body), &idData)
	if err != nil {
		return 0, err
	}
	if len(idData.Systems) == 0 {
		return 0, fmt.Errorf("system not found: %s", name)
	}

	id := idData.Systems[0].ID
	c.cacheMutex.Lock()
	c.systemIDCache[lowerName] = id
	c.cacheMutex.Unlock()

	return id, nil
}

// GetSystemName resolves an ID to its name, using caches first.
func (c *ESIClient) GetSystemName(id int) string {
	if id < 30000000 {
		return "Unknown"
	}
	idStr := strconv.Itoa(id)
	c.cacheMutex.RLock()
	if name, ok := c.systemNameCache[idStr]; ok {
		c.cacheMutex.RUnlock()
		return name
	}
	if name, ok := c.nameCache[id]; ok {
		c.cacheMutex.RUnlock()
		return name
	}
	c.cacheMutex.RUnlock()

	log.Printf("DEBUG: System ID %d not in cache, fetching from ESI.", id)
	nameMap, err := c.GetNames(context.Background(), []int{id})
	if err != nil {
		log.Printf("WARN: Failed to get name for ID %d: %v", id, err)
		return "Unknown"
	}
	name, ok := nameMap[id]
	if !ok {
		return "Unknown"
	}

	c.cacheMutex.Lock()
	c.nameCache[id] = name
	c.cacheMutex.Unlock()

	return name
}

// GetNames resolves a slice of IDs to a map of ID -> Name.
func (c *ESIClient) GetNames(ctx context.Context, ids []int) (map[int]string, error) {
	if len(ids) == 0 {
		return make(map[int]string), nil
	}
	var names []esiName
	body, _ := json.Marshal(ids)
	err := c.do(ctx, http.MethodPost, "/universe/names/", bytes.NewBuffer(body), &names)
	if err != nil {
		return nil, err
	}
	results := make(map[int]string)
	for _, nameEntry := range names {
		results[nameEntry.ID] = nameEntry.Name
	}
	return results, nil
}

// GetCharacterID resolves a character name to its ID.
func (c *ESIClient) GetCharacterID(ctx context.Context, name string) (int64, error) {
	var idData esiCharacterIDResult
	body, _ := json.Marshal([]string{name})
	err := c.do(ctx, http.MethodPost, "/universe/ids/", bytes.NewBuffer(body), &idData)
	if err != nil {
		return 0, err
	}
	if len(idData.Characters) == 0 {
		return 0, fmt.Errorf("character not found: %s", name)
	}
	return idData.Characters[0].ID, nil
}

// GetSystemDetails retrieves full system details from the local cache.
func (c *ESIClient) GetSystemDetails(id int) (*ESISystemInfo, error) {
	c.cacheMutex.RLock()
	defer c.cacheMutex.RUnlock()
	if sys, ok := c.systemInfoCache[id]; ok {
		return sys, nil
	}
	log.Printf("DEBUG: System ID %d was not found in the local info cache.", id)
	return nil, fmt.Errorf("system ID %d not found in cache", id)
}

/// Add this to your /internal/esi/client.go file

type EsiSystemKills struct {
	NpcKills  int `json:"npc_kills"`
	PodKills  int `json:"pod_kills"`
	ShipKills int `json:"ship_kills"`
	SystemID  int `json:"system_id"`
}

func (c *ESIClient) GetSystemKills(ctx context.Context) ([]EsiSystemKills, error) {
	var kills []EsiSystemKills
	// We need to add "/latest" because this is a versioned endpoint
	err := c.do(ctx, http.MethodGet, "/universe/system_kills/", nil, &kills)
	if err != nil {
		return nil, err
	}
	return kills, nil
}

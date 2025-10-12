package models

import "time"

// FrontendData is the main data structure passed to your templates.
type FrontendData struct {
	Connections []ConnectionInfo
	Leaderboard []LeaderboardEntry
	FeedbackURL string
	Path        []PathStep
}

type LeaderboardEntry struct {
	ScoutName string
	ScanCount int
}

// ConnectionInfo represents a single wormhole connection.
type ConnectionInfo struct {
	FromName    string
	ToName      string
	SignatureID string
	Eol         string
	Scout       string
	LastUpdated string
	EolStatus   string
}

type WingspanAPIResponse struct {
	Signatures map[string]Signature `json:"signatures"`
	Wormholes  map[string]Wormhole  `json:"wormholes"`
}

// Signature defines the structure for a single signature object.
type Signature struct {
	ID             string  `json:"id"`
	SignatureID    *string `json:"signatureID"` // Pointer to handle null values
	SystemID       string  `json:"systemID"`
	Type           string  `json:"type"`
	Name           *string `json:"name"` // Pointer to handle null values
	LifeTime       string  `json:"lifeTime"`
	LifeLeft       string  `json:"lifeLeft"`
	CreatedByID    string  `json:"createdByID"`
	CreatedByName  string  `json:"createdByName"`
	ModifiedByID   string  `json:"modifiedByID"`
	ModifiedByName string  `json:"modifiedByName"`
	ModifiedTime   string  `json:"modifiedTime"`
	MaskID         string  `json:"maskID"`
}

// Wormhole defines the structure for a single wormhole object.
type Wormhole struct {
	ID          string  `json:"id"`
	InitialID   string  `json:"initialID"`
	SecondaryID string  `json:"secondaryID"`
	Type        *string `json:"type"` // Pointer to handle null values
	Parent      string  `json:"parent"`
	Life        string  `json:"life"`
	Mass        string  `json:"mass"`
	MaskID      string  `json:"maskID"`
}

type TheraConnection struct {
	ID              string    `json:"id"`
	CreatedAt       time.Time `json:"created_at"`
	CreatedByID     int64     `json:"created_by_id"`
	CreatedByName   string    `json:"created_by_name"`
	UpdatedAt       time.Time `json:"updated_at"`
	UpdatedByID     int64     `json:"updated_by_id"`
	UpdatedByName   string    `json:"updated_by_name"`
	CompletedAt     time.Time `json:"completed_at"`
	CompletedByID   int64     `json:"completed_by_id"`
	CompletedByName string    `json:"completed_by_name"`
	Completed       bool      `json:"completed"`
	WhExitsOutward  bool      `json:"wh_exits_outward"`
	WhType          string    `json:"wh_type"`
	MaxShipSize     string    `json:"max_ship_size"`
	ExpiresAt       time.Time `json:"expires_at"`
	RemainingHours  int       `json:"remaining_hours"`
	SignatureType   string    `json:"signature_type"`
	OutSystemID     int       `json:"out_system_id"`
	OutSystemName   string    `json:"out_system_name"`
	OutSignature    string    `json:"out_signature"`
	InSystemID      int       `json:"in_system_id"`
	InSystemClass   string    `json:"in_system_class"`
	InSystemName    string    `json:"in_system_name"`
	InRegionID      int       `json:"in_region_id"`
	InRegionName    string    `json:"in_region_name"`
	InSignature     string    `json:"in_signature"`
}

// ... (Your other structs)

// PathStep represents one step in the calculated route.
type PathStep struct {
	SystemName     string
	JumpType       string
	SecurityStatus float64 // ADD THIS
	SecurityClass  string
	ShipKills      int // ADD THIS
	NpcKills       int // ADD THIS
}

type ESISystemInfo struct {
	Name           string  `json:"name"`
	SecurityStatus float64 `json:"security_status"`
	SystemID       int     `json:"system_id"`
}

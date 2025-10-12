package routing

import (
	"encoding/csv"
	"io"
	"os"
	"strconv"
)

type Edge struct {
	To     int
	Weight int
}

type Graph struct {
	staticAdj  map[int][]Edge // from CSV stargates
	dynamicAdj map[int][]Edge // from wormholes
}

func NewGraph() *Graph {
	return &Graph{
		staticAdj:  make(map[int][]Edge),
		dynamicAdj: make(map[int][]Edge),
	}
}

func (g *Graph) LoadCSV(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	r := csv.NewReader(f)
	// Optionally r.ReuseRecord = true if memory matters.
	// First line is header; read and discard if present.
	// Detect header by len and checking for non-integer fields.
	var isHeaderChecked bool

	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if !isHeaderChecked {
			isHeaderChecked = true
			// If the third field isn't an int, assume it's a header:
			if _, convErr := strconv.Atoi(rec[2]); convErr != nil {
				continue
			}
		}
		// Columns: fromRegionID, fromConstellationID, fromSolarSystemID,
		//          toSolarSystemID, toConstellationID, toRegionID
		fromSys, err1 := strconv.Atoi(rec[2])
		toSys, err2 := strconv.Atoi(rec[3])
		if err1 != nil || err2 != nil {
			continue
		}
		g.staticAdj[fromSys] = append(g.staticAdj[fromSys], Edge{To: toSys, Weight: 1})
		g.staticAdj[toSys] = append(g.staticAdj[toSys], Edge{To: fromSys, Weight: 1})
	}
	return nil
}

// ... (Edge and Graph structs)

// Clone creates a deep copy of the static graph for a single request.
func (g *Graph) Clone() *Graph {
	clone := NewGraph()
	// Copy the static connections
	for from, edges := range g.staticAdj {
		clone.staticAdj[from] = append(clone.staticAdj[from], edges...)
	}
	return clone
}

// ... (your existing Graph struct and functions)

// ADD THIS METHOD
// StaticAdjacencyListSize returns the number of systems in the static graph.
func (g *Graph) StaticAdjacencyListSize() int {
	return len(g.staticAdj)
}

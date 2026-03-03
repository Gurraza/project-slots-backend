package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"slots/internal/engine"
	"strconv"
)

// Game holds both the raw JSON for the frontend and the parsed logic for the backend.
type Game struct {
	Config *engine.GameConfig
}

// gameRegistry acts as a single source of truth for all loaded games.
var gameRegistry = make(map[string]Game)

// handleCORS sets standard headers to allow frontend access.
func handleCORS(w http.ResponseWriter) {
	// w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	// w.Header().Set("Access-Control-Allow-Origin", "http://192.168.68.102:5173")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

// ConfigEndpointHandler serves the raw JSON configuration to the frontend.
func ConfigEndpointHandler(w http.ResponseWriter, r *http.Request) {
	handleCORS(w)
	if r.Method == http.MethodOptions {
		return
	}

	gameID := r.URL.Query().Get("gameId")
	game, exists := gameRegistry[gameID]
	if !exists {
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	// Encoding the struct enforces your json tags
	json.NewEncoder(w).Encode(game.Config)
}

// PlayEndpointHandler executes game logic based on the requested gameId.
func PlayEndpointHandler(w http.ResponseWriter, r *http.Request) {
	handleCORS(w)
	if r.Method == http.MethodOptions {
		return
	}

	query := r.URL.Query()
	gameID := query.Get("gameId")
	seedStr := query.Get("seed")

	game, exists := gameRegistry[gameID]
	if !exists {
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	}

	var seed int64
	var err error

	if seedStr != "" {
		seed, err = strconv.ParseInt(seedStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid seed format", http.StatusBadRequest)
			return
		}
	} else {
		seed = rand.Int63n(9007199254740991)
	}
	timeline := game.Config.PlayGame(seed)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(timeline)
}

// loadGame is a helper to populate the registry at startup.
func loadGame(id string, filepath string) {
	rawJSON, err := os.ReadFile(filepath)
	if err != nil {
		panic(fmt.Sprintf("Failed to read %s: %v", filepath, err))
	}

	config, err := engine.LoadConfigFromJSON(rawJSON)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse %s: %v", filepath, err))
	}

	gameRegistry[id] = Game{
		Config: config,
	}
}

func main() {
	// 1. Initialize the registry
	loadGame("lines", "internal/games/lines.json")
	loadGame("clashofreels", "internal/games/clashofreels.json")

	// Setup CLI flags
	simFlag := flag.Bool("sim", false, "Run 1 million game simulation")
	gameFlag := flag.String("game", "lines", "Game ID to simulate")
	flag.Parse()

	if *simFlag {
		// Run simulation and exit. Assuming a baseline bet size of 1.0.
		RunSimulation(*gameFlag, 1_000_000)
		return
	}

	// 2. Register unified routes
	http.HandleFunc("/api/config", ConfigEndpointHandler)
	http.HandleFunc("/api/play", PlayEndpointHandler)

	fmt.Println("Go server running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

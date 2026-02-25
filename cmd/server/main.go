package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"slots/internal/engine"
	"strconv"
)

var linesGameConfig *engine.GameConfig
var clashofreelsConfig *engine.GameConfig

func PlayLinesEndpointHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
	timeline := linesGameConfig.PlayGame(rand.Int63())
	json.NewEncoder(w).Encode(timeline)
}

func PlayClashOfReelsEndpointHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
	// 1. Get query parameters
	query := r.URL.Query()
	seedStr := query.Get("seed")

	var s int64
	var err error

	// 2. Parse seed if provided, otherwise generate random fallback
	if seedStr != "" {
		s, err = strconv.ParseInt(seedStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid seed format", http.StatusBadRequest)
			return
		}
	} else {
		s = rand.Int63()
	}
	fmt.Println("Actual seed:" + strconv.FormatInt(s, 10))

	timeline := clashofreelsConfig.PlayGame(s)
	json.NewEncoder(w).Encode(timeline)
}

func main() {
	linesConfigJson, _ := os.ReadFile("internal/games/lines.json")
	clashofreelsJson, _ := os.ReadFile("internal/games/clashofreels.json")

	linesGameConfig, _ = engine.LoadConfigFromJSON(linesConfigJson)
	clashofreelsConfig, _ = engine.LoadConfigFromJSON(clashofreelsJson)

	fmt.Println("Go server running on http://localhost:8080")
	http.HandleFunc("/play/lines", PlayLinesEndpointHandler)
	http.HandleFunc("/play/clashofreels", PlayClashOfReelsEndpointHandler)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
	// for range 1 {
	// 	seed := rand.Int63()

	// 	fmt.Println("\nSeed:", seed)
	// 	// gameState := engine.NewGameState(linesGameConfig, seed)
	// 	// t := gameState.Spin()
	// 	t := linesGameConfig.PlayGame(seed)
	// 	models.Timeline(t).Print()
	// 	fmt.Println()
	// 	// encoder := json.NewEncoder(os.Stdout)
	// 	// encoder.SetIndent("", "  ") // Adds a 2-space indent
	// 	// encoder.Encode(t)
	// }
}

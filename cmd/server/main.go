package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"slots/internal/engine"
)

var linesGameConfig *engine.GameConfig

func PlayLinesEndpointHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	timeline := linesGameConfig.PlayGame(rand.Int63())
	json.NewEncoder(w).Encode(timeline)
}

func main() {
	linesConfigPath := "internal/games/lines.json"
	linesConfigJson, err := os.ReadFile(linesConfigPath)
	if err != nil {
		log.Fatalf("Critical Error: Could not read config file at %s: %v", linesConfigPath, err)
	}

	linesGameConfig, err = engine.LoadConfigFromJSON(linesConfigJson)
	if err != nil {
		log.Fatalf("Critical Error: Failed to parse JSON config: %v", err)
	}

	fmt.Println("Go server running on http://localhost:8080")
	http.HandleFunc("/play/lines", PlayLinesEndpointHandler)

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

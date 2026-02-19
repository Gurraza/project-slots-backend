package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"slots/internal/engine"
	"slots/internal/timeline"
)

var linesGameConfig *engine.GameConfig

// func PlayLinesEndpointHandler(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
// 	timeline := PlayLines()
// 	json.NewEncoder(w).Encode(timeline)
// }

func main() {
	linesConfigPath := "internal/games/lines.config"
	linesConfigJson, err := os.ReadFile(linesConfigPath)
	if err != nil {
		log.Fatalf("Critical Error: Could not read config file at %s: %v", linesConfigPath, err)
	}

	c, err := engine.LoadConfigFromJSON(linesConfigJson)
	if err != nil {
		log.Fatalf("Critical Error: Failed to parse JSON config: %v", err)
	}

	linesGameConfig = c

	// fmt.Println("Go server running on http://localhost:8080")
	// http.HandleFunc("/api/play/lines", PlayLinesEndpointHandler)

	// if err := http.ListenAndServe(":8080", nil); err != nil {
	// 	panic(err)
	// }
	for range 1 {
		seed := rand.Int63()
		fmt.Println("\nSeed:", seed)
		gameState := engine.NewGameState(linesGameConfig, seed)
		t := gameState.PlayRound()
		PrintTimeline(t)
		fmt.Println()
	}
}
func PrintTimeline(t []*timeline.TimelineEvent) {
	fmt.Printf("Events: %d Total Win Amount %.2f \n", len(t), t[len(t)-1].TotalWinAmount)
	for i, e := range t {
		fmt.Printf("Grid %d type "+e.Type+" win amount %.2f\n", i, e.WinAmount)

		if e.Meta != nil {
			fmt.Printf("Lines %+V\n", e.Meta)
		}

		g := e.GridSnapshot.Cells
		if len(g) == 0 {
			continue
		}

		cols := len(g)
		rows := len(g[0])

		// 1. Loop through rows (Y axis) first
		for y := 0; y < rows; y++ {
			// 2. Loop through columns (X axis) for the current row
			for x := 0; x < cols; x++ {
				// Use %3d to pad single-digit numbers so columns align perfectly
				fmt.Printf("%3d ", g[x][y])
			}
			// 3. Print a newline at the end of each row
			fmt.Println()
		}
		fmt.Println()
	}
}

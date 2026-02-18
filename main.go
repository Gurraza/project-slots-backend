package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
)

type RandomNumberResponse struct {
	Value int `json:"value"`
}

func randomNumberHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")

	var seed int64 = 64
	source := rand.NewSource(seed)
	rng := rand.New(source)
	num := rng.Intn(10) + 1

	response := RandomNumberResponse{Value: num}
	json.NewEncoder(w).Encode(response)
}

func playHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	seed := rand.Int63()
	gameState := NewGameState(&GameConfig{Cols: 5, Rows: 3, Symbols: []SymbolDef{}}, seed)
	gameState.Features = append(gameState.Features, NewPaylineFeature(gameState))
	var timeline []TimelineEvent = PlayRound(gameState)
	fmt.Println(timeline)
	json.NewEncoder(w).Encode(timeline)
}

func main() {
	http.HandleFunc("/api/random", randomNumberHandler)
	http.HandleFunc("/api/play", playHandler)

	fmt.Println("Go server running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

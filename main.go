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

func PlayLinesEndpointHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	timeline := PlayLines()
	json.NewEncoder(w).Encode(timeline)
}

func main() {
	http.HandleFunc("/api/random", randomNumberHandler)
	http.HandleFunc("/api/play/lines", PlayLinesEndpointHandler)

	fmt.Println("Go server running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
	// timeline := PlayLines()
	// data, _ := json.MarshalIndent(timeline, "", " ")
	// fmt.Println(string(data))
}

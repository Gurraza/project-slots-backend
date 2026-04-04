// package main

// import (
// 	"fmt"
// 	"math/rand"
// 	"sort"
// 	"sync"
// 	"sync/atomic"
// 	"time"
// )

// type SeedResult struct {
// 	Seed    float64 // Using float64 for Win/Length values
// 	Value   float64
// 	RawSeed int64
// }

// type SimResults struct {
// 	TotalSpins int64
// 	TotalWin   float64
// 	MaxWin     float64
// 	WinCounts  int64
// 	Buckets    map[string]int64

// 	// Top 3 Tracking
// 	TopWins    []SeedResult
// 	TopLengths []SeedResult

// 	Mu sync.Mutex
// }

// func RunSimulation(gameID string, iterations int) {
// 	game, exists := gameRegistry[gameID]
// 	if !exists {
// 		fmt.Printf("Game %s not found\n", gameID)
// 		return
// 	}

// 	results := &SimResults{
// 		TotalSpins: int64(iterations),
// 		Buckets:    make(map[string]int64),
// 		TopWins:    make([]SeedResult, 0, 3),
// 		TopLengths: make([]SeedResult, 0, 3),
// 	}

// 	start := time.Now()
// 	jobs := make(chan int, 1000)
// 	var wg sync.WaitGroup
// 	var completed int64 // For progress tracking

// 	workerCount := 12
// 	checkpoint := int64(iterations / 10)

// 	for w := 0; w < workerCount; w++ {
// 		wg.Add(1)
// 		go func() {
// 			defer wg.Done()
// 			for range jobs {
// 				seed := rand.Int63n(9007199254740991)
// 				timeline := game.Config.PlayGame(seed)

// 				var win float64
// 				length := float64(len(timeline))
// 				if length > 0 {
// 					win = timeline[len(timeline)-1].TotalWinAmount
// 				}

// 				results.Mu.Lock()
// 				// 1. Basic Stats
// 				results.TotalWin += win
// 				if win > 0 {
// 					results.WinCounts++
// 					updateBuckets(results, win)
// 				}

// 				// 2. Track Top 3 Wins
// 				updateTopList(&results.TopWins, SeedResult{Value: win, RawSeed: seed})

// 				// 3. Track Top 3 Lengths
// 				updateTopList(&results.TopLengths, SeedResult{Value: length, RawSeed: seed})

// 				results.Mu.Unlock()

// 				// 4. Progress Reporting
// 				newCompleted := atomic.AddInt64(&completed, 1)
// 				if newCompleted%checkpoint == 0 {
// 					fmt.Printf("Progress: %d%%\n", (newCompleted*100)/int64(iterations))
// 				}
// 			}
// 		}()
// 	}

// 	for i := 0; i < iterations; i++ {
// 		jobs <- i
// 	}
// 	close(jobs)
// 	wg.Wait()

// 	printDetailedReport(gameID, results, time.Since(start))
// }

// // Helper to maintain a sorted top 3 list
// func updateTopList(list *[]SeedResult, entry SeedResult) {
// 	*list = append(*list, entry)
// 	sort.Slice(*list, func(i, j int) bool {
// 		return (*list)[i].Value > (*list)[j].Value
// 	})
// 	if len(*list) > 3 {
// 		*list = (*list)[:3]
// 	}
// }

// func updateBuckets(r *SimResults, win float64) {
// 	if win >= 100 {
// 		r.Buckets["100x+"]++
// 	} else if win >= 50 {
// 		r.Buckets["50x-100x"]++
// 	} else if win >= 10 {
// 		r.Buckets["10x-50x"]++
// 	} else {
// 		r.Buckets["<10x"]++
// 	}
// }
// func printDetailedReport(id string, r *SimResults, d time.Duration) {
// 	fmt.Printf("\n--- SIMULATION COMPLETE: %s (%v) ---\n", id, d)
// 	// fmt.Printf("RTP: %.2f%% | Hit Freq: %.2f%%\n", (r.TotalWin/float64(r.TotalSpins))*100, (float64(r.WinCounts)/float64(r.TotalSpins))*100)

// 	rtp := (r.TotalWin / float64(r.TotalSpins)) * 100
// 	hitFreq := (float64(r.WinCounts) / float64(r.TotalSpins)) * 100

// 	fmt.Println("-------------------------------------------")
// 	fmt.Printf("SIMULATION REPORT: %s\n", id)
// 	fmt.Printf("Spins:      %d\n", r.TotalSpins)
// 	fmt.Printf("Time:       %v\n", d)
// 	fmt.Printf("RTP:        %.2f%%\n", rtp)
// 	fmt.Printf("Hit Freq:   %.2f%%\n", hitFreq)
// 	fmt.Printf("Max Win:    %.2f x\n", r.MaxWin)
// 	fmt.Println("WIN DISTRIBUTION:")
// 	for label, count := range r.Buckets {
// 		fmt.Printf("  %s: %d\n", label, count)
// 	}
// 	fmt.Println("-------------------------------------------")

// 	fmt.Println("\nTOP 3 HIGHEST WINS:")
// 	for i, res := range r.TopWins {
// 		fmt.Printf(" %d. Win: %.2f | Seed: %d\n", i+1, res.Value, res.RawSeed)
// 	}

// 	fmt.Println("\nTOP 3 LONGEST TIMELINES:")
// 	for i, res := range r.TopLengths {
// 		fmt.Printf(" %d. Events: %.0f | Seed: %d\n", i+1, res.Value, res.RawSeed)
// 	}
// }

// func printReport(id string, r *SimResults, duration time.Duration) {
// }

package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

type SeedResult struct {
	Seed int64   `json:"seed"`
	Win  float64 `json:"win"`
	Type string  `json:"type"`
}

func RunSimulation(gameID string) {
	game, exists := gameRegistry[gameID]
	if !exists {
		fmt.Printf("Game %s not found\n", gameID)
		return
	}

	// Målparametrar för de 100 rundorna
	// Updated Target Distribution for a deeper net loss but bigger highlights
	const betAmount = 1.0
	const targetLosses = 55 // Increased from 40
	const targetLDWs = 35   // Decreased from 45
	const targetWins = 10   // Decreased from 15

	// Set your target Return. If Cost is 100, Return of 30-40 means Net Result of -60 to -70.
	const minTotalReturn = 90.0
	const maxTotalReturn = 145.0

	// Pooler för att spara giltiga kandidater
	var poolLosses []SeedResult
	var poolLDWs []SeedResult
	var poolWins []SeedResult

	var mu sync.Mutex
	var wg sync.WaitGroup
	var iterations int64

	start := time.Now()
	workerCount := 12
	jobs := make(chan int, 1000)
	done := make(chan bool)

	// Fas 1: Fyll poolerna med kandidater
	for w := 0; w < workerCount; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range jobs {
				currentIter := atomic.AddInt64(&iterations, 1)

				if currentIter%10000 == 0 {
					mu.Lock()
					fmt.Printf("Samlar in... Iter: %d | Pool Losses: %d | Pool LDWs: %d | Pool Wins: %d\n",
						currentIter, len(poolLosses), len(poolLDWs), len(poolWins))
					mu.Unlock()
				}

				seed := rand.Int63n(9007199254740991)
				timeline := game.Config.PlayGame(seed)

				var win float64
				if len(timeline) > 0 {
					win = timeline[len(timeline)-1].TotalWinAmount
				}

				mu.Lock()
				// Samla in minst det dubbla antalet som behövs för att ha urval att pussla med
				isComplete := len(poolLosses) >= targetLosses*2 && len(poolLDWs) >= targetLDWs*2 && len(poolWins) >= targetWins*5

				if !isComplete {
					if win == 0 && len(poolLosses) < targetLosses*2 {
						poolLosses = append(poolLosses, SeedResult{Seed: seed, Win: win, Type: "LOSS"})
					} else if win > 0 && win < betAmount && len(poolLDWs) < targetLDWs*2 {
						poolLDWs = append(poolLDWs, SeedResult{Seed: seed, Win: win, Type: "LDW"})
					} else if win >= betAmount && win <= 40.0 && len(poolWins) < targetWins*5 {
						// Increased upper limit from 15.0 to 40.0 to capture Grand Wins
						poolWins = append(poolWins, SeedResult{Seed: seed, Win: win, Type: "WIN"})
					}
				}
				mu.Unlock()

				if isComplete {
					select {
					case <-done:
					default:
						close(done)
					}
					return
				}
			}
		}()
	}

	go func() {
		i := 0
		for {
			select {
			case <-done:
				close(jobs)
				return
			default:
				jobs <- i
				i++
			}
		}
	}()

	wg.Wait()
	fmt.Println("\nPooler fyllda. Letar efter en sekvens som matchar målsättningen för nettovinst...")

	// Fas 2: Hitta en kombination som ger önskat nettoresultat
	rand.Seed(time.Now().UnixNano())
	var finalSequence []SeedResult
	var totalWin float64
	foundValidSequence := false

	for attempts := 0; attempts < 100000; attempts++ {
		totalWin = 0.0
		finalSequence = make([]SeedResult, 0, 100)

		// Välj slumpmässigt från poolerna
		rand.Shuffle(len(poolLosses), func(i, j int) { poolLosses[i], poolLosses[j] = poolLosses[j], poolLosses[i] })
		rand.Shuffle(len(poolLDWs), func(i, j int) { poolLDWs[i], poolLDWs[j] = poolLDWs[j], poolLDWs[i] })
		rand.Shuffle(len(poolWins), func(i, j int) { poolWins[i], poolWins[j] = poolWins[j], poolWins[i] })

		selectedLosses := poolLosses[:targetLosses]
		selectedLDWs := poolLDWs[:targetLDWs]
		selectedWins := poolWins[:targetWins]

		finalSequence = append(finalSequence, selectedLosses...)
		finalSequence = append(finalSequence, selectedLDWs...)
		finalSequence = append(finalSequence, selectedWins...)

		for _, s := range finalSequence {
			totalWin += s.Win
		}

		// Validera om summan är inom intervallet
		if totalWin >= minTotalReturn && totalWin <= maxTotalReturn {
			foundValidSequence = true
			break
		}
	}

	if !foundValidSequence {
		fmt.Println("Kunde inte hitta en sekvens som matchar villkoren. Justera gränserna eller kör igen.")
		return
	}

	// Blanda den slutgiltiga sekvensen
	rand.Shuffle(len(finalSequence), func(i, j int) {
		finalSequence[i], finalSequence[j] = finalSequence[j], finalSequence[i]
	})

	fmt.Printf("\n--- SEKVENSSAMMANSTÄLLNING (%v) ---\n", time.Since(start))
	for i, s := range finalSequence {
		if s.Type != "LOSS" {
			fmt.Printf("Snurr %03d | %-4s | Win: %.2f\n", i+1, s.Type, s.Win)
		}
	}

	fmt.Printf("\nTotalt antal snurr: %d\n", len(finalSequence))
	fmt.Printf("Total Kostnad: %.2f kr\n", float64(len(finalSequence))*betAmount)
	fmt.Printf("Total Vinst (Return): %.2f kr\n", totalWin)
	fmt.Printf("Nettoresultat: %.2f kr\n\n", totalWin-(float64(len(finalSequence))*betAmount))

	var rawSeeds []int64
	for _, s := range finalSequence {
		rawSeeds = append(rawSeeds, s.Seed)
	}

	jsonOutput, _ := json.Marshal(rawSeeds)
	fmt.Println("JS Array (Kopiera till frontend):")
	fmt.Println(string(jsonOutput))
}

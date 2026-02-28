package main

import (
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type SeedResult struct {
	Seed    float64 // Using float64 for Win/Length values
	Value   float64
	RawSeed int64
}

type SimResults struct {
	TotalSpins int64
	TotalWin   float64
	MaxWin     float64
	WinCounts  int64
	Buckets    map[string]int64

	// Top 3 Tracking
	TopWins    []SeedResult
	TopLengths []SeedResult

	Mu sync.Mutex
}

func RunSimulation(gameID string, iterations int) {
	game, exists := gameRegistry[gameID]
	if !exists {
		fmt.Printf("Game %s not found\n", gameID)
		return
	}

	results := &SimResults{
		TotalSpins: int64(iterations),
		Buckets:    make(map[string]int64),
		TopWins:    make([]SeedResult, 0, 3),
		TopLengths: make([]SeedResult, 0, 3),
	}

	start := time.Now()
	jobs := make(chan int, 1000)
	var wg sync.WaitGroup
	var completed int64 // For progress tracking

	workerCount := 12
	checkpoint := int64(iterations / 10)

	for w := 0; w < workerCount; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range jobs {
				seed := rand.Int63()
				timeline := game.Config.PlayGame(seed)

				var win float64
				length := float64(len(timeline))
				if length > 0 {
					win = timeline[len(timeline)-1].TotalWinAmount
				}

				results.Mu.Lock()
				// 1. Basic Stats
				results.TotalWin += win
				if win > 0 {
					results.WinCounts++
					updateBuckets(results, win)
				}

				// 2. Track Top 3 Wins
				updateTopList(&results.TopWins, SeedResult{Value: win, RawSeed: seed})

				// 3. Track Top 3 Lengths
				updateTopList(&results.TopLengths, SeedResult{Value: length, RawSeed: seed})

				results.Mu.Unlock()

				// 4. Progress Reporting
				newCompleted := atomic.AddInt64(&completed, 1)
				if newCompleted%checkpoint == 0 {
					fmt.Printf("Progress: %d%%\n", (newCompleted*100)/int64(iterations))
				}
			}
		}()
	}

	for i := 0; i < iterations; i++ {
		jobs <- i
	}
	close(jobs)
	wg.Wait()

	printDetailedReport(gameID, results, time.Since(start))
}

// Helper to maintain a sorted top 3 list
func updateTopList(list *[]SeedResult, entry SeedResult) {
	*list = append(*list, entry)
	sort.Slice(*list, func(i, j int) bool {
		return (*list)[i].Value > (*list)[j].Value
	})
	if len(*list) > 3 {
		*list = (*list)[:3]
	}
}

func updateBuckets(r *SimResults, win float64) {
	if win >= 100 {
		r.Buckets["100x+"]++
	} else if win >= 50 {
		r.Buckets["50x-100x"]++
	} else if win >= 10 {
		r.Buckets["10x-50x"]++
	} else {
		r.Buckets["<10x"]++
	}
}
func printDetailedReport(id string, r *SimResults, d time.Duration) {
	fmt.Printf("\n--- SIMULATION COMPLETE: %s (%v) ---\n", id, d)
	// fmt.Printf("RTP: %.2f%% | Hit Freq: %.2f%%\n", (r.TotalWin/float64(r.TotalSpins))*100, (float64(r.WinCounts)/float64(r.TotalSpins))*100)

	rtp := (r.TotalWin / float64(r.TotalSpins)) * 100
	hitFreq := (float64(r.WinCounts) / float64(r.TotalSpins)) * 100

	fmt.Println("-------------------------------------------")
	fmt.Printf("SIMULATION REPORT: %s\n", id)
	fmt.Printf("Spins:      %d\n", r.TotalSpins)
	fmt.Printf("Time:       %v\n", d)
	fmt.Printf("RTP:        %.2f%%\n", rtp)
	fmt.Printf("Hit Freq:   %.2f%%\n", hitFreq)
	fmt.Printf("Max Win:    %.2f x\n", r.MaxWin)
	fmt.Println("WIN DISTRIBUTION:")
	for label, count := range r.Buckets {
		fmt.Printf("  %s: %d\n", label, count)
	}
	fmt.Println("-------------------------------------------")

	fmt.Println("\nTOP 3 HIGHEST WINS:")
	for i, res := range r.TopWins {
		fmt.Printf(" %d. Win: %.2f | Seed: %d\n", i+1, res.Value, res.RawSeed)
	}

	fmt.Println("\nTOP 3 LONGEST TIMELINES:")
	for i, res := range r.TopLengths {
		fmt.Printf(" %d. Events: %.0f | Seed: %d\n", i+1, res.Value, res.RawSeed)
	}
}

func printReport(id string, r *SimResults, duration time.Duration) {
}

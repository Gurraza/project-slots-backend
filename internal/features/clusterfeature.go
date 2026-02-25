package features

import (
	"slots/internal/models"
)

type ClusterFeature struct {
	BaseFeature
	ExcludeSymbolIDs []int
	ClusterSize      int
}

type Cluster struct {
	Points []ExplosionPoint
	Symbol models.SymbolDef
}

type ExplosionPoint struct {
	X             int `json:"x"`
	Y             int `json:"y"`
	ReplacementID int `json:"replacementId"`
}

func NewClusterFeature(clusterSize int, excludeSymbolIds []int) *ClusterFeature {
	return &ClusterFeature{
		BaseFeature: BaseFeature{
			Type: "CLUSTER_FEATURE",
		},
		ClusterSize:      clusterSize,
		ExcludeSymbolIDs: excludeSymbolIds,
	}
}
func (f *ClusterFeature) OnGridIdle(ctx FeatureContext) bool {
	clusters := FindClusters(ctx.GetGrid(), ctx.GetSymbols(), f.ClusterSize)
	if len(clusters) == 0 {
		return false
	}

	var allPoints []models.Point
	var allReplacements []int
	var mergedExplosions []ExplosionPoint

	for i := range clusters {
		for j := range clusters[i].Points {
			px := clusters[i].Points[j].X
			py := clusters[i].Points[j].Y

			// Generate replacement using correct spatial coordinates
			repSymbol := ctx.GetRandomSymbol(ctx.GetGrid(), px, py)

			// Map replacement ID directly to the point
			clusters[i].Points[j].ReplacementID = repSymbol.ID

			// Populate slices for legacy models.ExplodeAndCascade signature and Meta payload
			allPoints = append(allPoints, models.Point{X: px, Y: py})
			allReplacements = append(allReplacements, repSymbol.ID)
			mergedExplosions = append(mergedExplosions, clusters[i].Points[j])
		}
	}

	ctx.SetGrid(models.ExplodeAndCascade(ctx.GetGrid(), allPoints, allReplacements))

	spinWinAmount := CalculatePayout(clusters)

	ctx.PushTimeline(&models.TimelineEvent{
		Type:         "EXPLODE_AND_CASCADE_FEATURE",
		GridSnapshot: ctx.GetGrid().Copy(),
		WinAmount:    spinWinAmount,
		Meta: map[string]interface{}{
			// Send the unified structure to the frontend
			"explosions": mergedExplosions,
		},
	})
	return true
}

// func (f *ClusterFeature) OnGridIdle(ctx FeatureContext) bool {
// 	clusters := FindClusters(ctx.GetGrid(), ctx.GetSymbols(), f.ClusterSize)
// 	mergedCluster := make([]models.Point, 0)
// 	for _, c := range clusters {
// 		mergedCluster = append(mergedCluster, c.Points...)
// 	}
// 	if len(mergedCluster) == 0 {
// 		return false
// 	}

// 	replacements := make([]int, len(mergedCluster))
// 	for i := range replacements {
// 		replacements[i] = ctx.GetRandomSymbol(ctx.GetGrid(), i, 0).ID
// 	}

// 	ctx.SetGrid(models.ExplodeAndCascade(ctx.GetGrid(), mergedCluster, replacements))

// 	spinWinAmount := CalculatePayout(clusters)

//		ctx.PushTimeline(&models.TimelineEvent{
//			Type:         "EXPLODE_AND_CASCADE_FEATURE",
//			GridSnapshot: ctx.GetGrid().Copy(),
//			WinAmount:    spinWinAmount,
//			Meta: map[string]interface{}{
//				"points":       mergedCluster,
//				"replacements": replacements,
//			},
//		})
//		return true
//	}
func FindClusters(g *models.Grid, symbols map[int]*models.SymbolDef, clusterSize int) []Cluster {
	defMap := make(map[int]*models.SymbolDef)
	for _, d := range symbols {
		defMap[d.ID] = d
	}

	isWild := func(id int) bool {
		def, exists := defMap[id]
		return exists && len(def.MatchesWith) > 0 && def.MatchesWith[0] == "*"
	}

	areCompatible := func(baseID, targetID int) bool {
		if baseID == targetID {
			return true
		}
		baseDef, baseExists := defMap[baseID]
		targetDef, targetExists := defMap[targetID]
		if !baseExists || !targetExists {
			return false
		}
		for _, m := range baseDef.MatchesWith {
			if m == "*" || m == targetDef.Name {
				return true
			}
		}
		for _, m := range targetDef.MatchesWith {
			if m == "*" || m == baseDef.Name {
				return true
			}
		}
		return false
	}

	// Global lock for all symbols (Standard and Wilds)
	visited := make([][]bool, g.Cols)
	for x := 0; x < g.Cols; x++ {
		visited[x] = make([]bool, g.Rows)
	}

	var clusters []Cluster

	// Scan left-to-right, top-to-bottom
	for x := 0; x < g.Cols; x++ {
		for y := 0; y < g.Rows; y++ {
			startID := g.Cells[x][y]

			// Skip if already claimed by a previous cluster, or if it is a Wild
			if visited[x][y] || isWild(startID) {
				continue
			}

			localVisited := make(map[models.Point]bool)
			queue := []models.Point{{X: x, Y: y}}
			localVisited[models.Point{X: x, Y: y}] = true

			var currentCluster []ExplosionPoint
			clusterSymbolID := startID

			for len(queue) > 0 {
				curr := queue[0]
				queue = queue[1:]

				currentCluster = append(currentCluster, ExplosionPoint{X: curr.X, Y: curr.Y})

				dirs := []models.Point{{X: 0, Y: -1}, {X: 0, Y: 1}, {X: -1, Y: 0}, {X: 1, Y: 0}}
				for _, d := range dirs {
					nx, ny := curr.X+d.X, curr.Y+d.Y

					if nx >= 0 && nx < g.Cols && ny >= 0 && ny < g.Rows {
						np := models.Point{X: nx, Y: ny}

						// Check local AND global visited state
						if !localVisited[np] && !visited[nx][ny] {
							neighborID := g.Cells[nx][ny]

							if areCompatible(clusterSymbolID, neighborID) {
								localVisited[np] = true
								queue = append(queue, np)
							}
						}
					}
				}
			}

			if len(currentCluster) >= clusterSize {
				// Lock ALL points in the cluster (First Claim)
				for _, p := range currentCluster {
					visited[p.X][p.Y] = true
				}

				clusters = append(clusters, Cluster{
					Points: currentCluster,
					Symbol: *defMap[clusterSymbolID],
				})
			} else {
				// Lock the failed root to prevent redundant BFS cycles
				visited[x][y] = true
			}
		}
	}

	return clusters
}

// func FindClusters(g *models.Grid, symbols map[int]*models.SymbolDef, clusterSize int) []Cluster {
// 	// 1. Build a lookup map for efficient SymbolDef access by ID
// 	defMap := make(map[int]*models.SymbolDef)
// 	for _, d := range symbols {
// 		defMap[d.ID] = d
// 	}

// 	// 2. Initialize visited matrix to keep track of processed cells
// 	visited := make([][]bool, g.Cols)
// 	for x := 0; x < g.Cols; x++ {
// 		visited[x] = make([]bool, g.Rows)
// 	}

// 	var clusters []Cluster

// 	// Helper function to check if two IDs are compatible
// 	areCompatible := func(id1, id2 int) bool {
// 		// Exact match is always a connection
// 		if id1 == id2 {
// 			return true
// 		}

// 		def1, exists1 := defMap[id1]
// 		def2, exists2 := defMap[id2]

// 		// If either definition is missing, we rely solely on exact ID match (checked above)
// 		if !exists1 || !exists2 {
// 			return false
// 		}

// 		// Check if def1 matches def2 (by specific name or wildcard)
// 		for _, matchName := range def1.MatchesWith {
// 			if matchName == "*" || matchName == def2.Name {
// 				return true
// 			}
// 		}

// 		// Check if def2 matches def1 (symmetric check)
// 		for _, matchName := range def2.MatchesWith {
// 			if matchName == "*" || matchName == def1.Name {
// 				return true
// 			}
// 		}

// 		return false
// 	}

// 	// 3. Iterate through every cell in the grid
// 	for x := 0; x < g.Cols; x++ {
// 		for y := 0; y < g.Rows; y++ {
// 			// If already visited, skip
// 			if visited[x][y] {
// 				continue
// 			}

// 			// Start a new cluster
// 			currentCluster := []models.Point{}

// 			// Queue for BFS
// 			queue := []models.Point{{X: x, Y: y}}
// 			visited[x][y] = true

// 			// Perform Flood Fill (BFS)
// 			for len(queue) > 0 {
// 				curr := queue[0]
// 				queue = queue[1:] // Dequeue

// 				// Add current point to the cluster
// 				currentCluster = append(currentCluster, curr)

// 				// Define 4-way directions (Up, Down, Left, Right)
// 				dirs := []models.Point{{X: 0, Y: -1}, {X: 0, Y: 1}, {X: -1, Y: 0}, {X: 1, Y: 0}}

// 				for _, d := range dirs {
// 					nx, ny := curr.X+d.X, curr.Y+d.Y

// 					// Boundary Check
// 					if nx >= 0 && nx < g.Cols && ny >= 0 && ny < g.Rows {
// 						if !visited[nx][ny] {
// 							currentVal := g.Cells[curr.X][curr.Y]
// 							neighborVal := g.Cells[nx][ny]

// 							// Compatibility Check
// 							if areCompatible(currentVal, neighborVal) {
// 								visited[nx][ny] = true
// 								queue = append(queue, models.Point{X: nx, Y: ny})
// 							}
// 						}
// 					}
// 				}
// 			}

// 			if len(currentCluster) >= clusterSize {
// 				dominantSymbolID := g.Cells[x][y]

// 				// Find the actual paying symbol, ignoring Wilds
// 				for _, p := range currentCluster {
// 					id := g.Cells[p.X][p.Y]
// 					def := defMap[id]

// 					// Check if the current dominant is a Wild (assuming Wilds have "*" in MatchesWith)
// 					isCurrentWild := len(defMap[dominantSymbolID].MatchesWith) > 0 && defMap[dominantSymbolID].MatchesWith[0] == "*"
// 					isNewWild := len(def.MatchesWith) > 0 && def.MatchesWith[0] == "*"

// 					if isCurrentWild && !isNewWild {
// 						dominantSymbolID = id
// 						break // Found the paying symbol
// 					}
// 				}

// 				clusters = append(clusters, Cluster{
// 					Points: currentCluster,
// 					Symbol: *defMap[dominantSymbolID],
// 				})
// 			}
// 		}
// 	}

// 	return clusters
// }

// CalculatePayout takes a slice of clusters and returns the total float64 win amount.
func CalculatePayout(clusters []Cluster) float64 {
	var totalWin float64

	for _, cluster := range clusters {
		count := len(cluster.Points)
		if count == 0 {
			continue
		}

		payouts := cluster.Symbol.Payouts
		if len(payouts) == 0 {
			continue // Failsafe: No payouts defined for this symbol
		}

		// The Payouts array is 0-indexed.
		// Index 0 = 1 symbol, Index 1 = 2 symbols, Index 2 = 3 symbols, etc.
		payoutIndex := count - 1

		// Cap the index at the maximum defined payout.
		// If a player gets a cluster of 15, but the array only goes up to 5,
		// they receive the maximum payout defined at the end of the array.
		if payoutIndex >= len(payouts) {
			payoutIndex = len(payouts) - 1
		}

		totalWin += payouts[payoutIndex]
	}

	return totalWin
}

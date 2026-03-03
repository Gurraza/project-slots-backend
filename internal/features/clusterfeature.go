package features

import (
	"slots/internal/models"
	"sort"
)

type ClusterFeature struct {
	BaseFeature
	ExcludeSymbolIDs []int
	ClusterSize      int
}

type Cluster struct {
	Points []models.Point
	Symbol models.SymbolDef
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
	var mergedPoints []models.Point

	for _, cluster := range clusters {
		for _, point := range cluster.Points {
			mergedPoints = append(mergedPoints, point)
		}
	}

	sort.Slice(mergedPoints, func(i, j int) bool {
		if mergedPoints[i].X == mergedPoints[j].X {
			return mergedPoints[i].Y < mergedPoints[j].Y
		}
		return mergedPoints[i].X < mergedPoints[j].X
	})

	explosions := ctx.GetGrid().ExplodeAndCascade(mergedPoints, func(x, y int) *models.SymbolDef {
		return ctx.GetRandomSymbol(ctx.GetGrid(), x, y)
	})

	spinWinAmount := CalculatePayout(clusters)

	ctx.PushTimeline(&models.TimelineEvent{
		Type:         "EXPLODE_AND_CASCADE_FEATURE",
		GridSnapshot: ctx.GetGrid().Copy(),
		WinAmount:    spinWinAmount,
		Meta: map[string]interface{}{
			// Send the unified structure to the frontend
			"explosions": explosions,
		},
	})
	return true
}

// Takes a grid and symbolmap where id points to its definitoin and a clusterSize.
// Returns a slice of clusters where each cluster has a slice of points and a symbolId of the cluster
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

			var currentCluster []models.Point
			clusterSymbolID := startID

			for len(queue) > 0 {
				curr := queue[0]
				queue = queue[1:]

				currentCluster = append(currentCluster, models.Point{X: curr.X, Y: curr.Y})

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

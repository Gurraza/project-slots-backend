package features

import (
	"slots/internal/models"
)

type ClusterFeature struct {
	BaseFeature
	ExcludeSymbolIDs []int
	ClusterSize      int
}

func NewClusterFeature(prio int, clusterSize int, excludeSymbolIds []int) *ClusterFeature {
	return &ClusterFeature{
		BaseFeature: BaseFeature{
			Type:     "CLUSTER_FEATURE",
			Priority: prio,
		},
		ClusterSize:      clusterSize,
		ExcludeSymbolIDs: excludeSymbolIds,
	}
}

func (f *ClusterFeature) OnGridIdle(ctx FeatureContext) bool {
	clusters := ctx.GetGrid().FindClusters(ctx.GetSymbols(), f.ClusterSize)

	if len(clusters) == 0 {
		return false
	}

	merged := models.MergeClusterPoints(clusters)

	explosions := ctx.GetGrid().ExplodeAndCascade(merged, func(x, y int) *models.SymbolDef {
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

// CalculatePayout takes a slice of clusters and returns the total float64 win amount.
func CalculatePayout(clusters []models.Cluster) float64 {
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

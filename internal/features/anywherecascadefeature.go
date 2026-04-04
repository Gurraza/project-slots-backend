package features

import (
	"slots/internal/models"
)

type ScatterCascadeFeature struct {
	BaseFeature
	TargetSymbolIDs []int // Specific symbols to track. Leave empty to evaluate all symbols.
	MinCount        int   // Minimum occurrences required anywhere on the grid to trigger explosion.
}

func NewAnywhereCascadeFeature(prio int, minCount int, targetSymbolIDs []int) *ScatterCascadeFeature {
	return &ScatterCascadeFeature{
		BaseFeature: BaseFeature{
			Type:     "ANYWHERE_CASCADE_FEATURE",
			Priority: prio,
		},
		MinCount:        minCount,
		TargetSymbolIDs: targetSymbolIDs,
	}
}

func (f *ScatterCascadeFeature) OnGridIdle(ctx FeatureContext) bool {
	grid := ctx.GetGrid()
	symbols := ctx.GetSymbols()

	// 1. Count all symbol occurrences and record their coordinates
	counts := make(map[int][]models.Point)
	for x := range grid.Cols {
		for y := range grid.Rows {
			id := grid.Get(x, y)
			counts[id] = append(counts[id], models.Point{X: x, Y: y})
		}
	}

	var toExplode []models.Point
	var totalWin float64

	// 2. Establish valid target symbols
	validTargets := make(map[int]bool)
	if len(f.TargetSymbolIDs) > 0 {
		for _, id := range f.TargetSymbolIDs {
			validTargets[id] = true
		}
	} else {
		for id := range symbols {
			validTargets[id] = true
		}
	}

	// 3. Evaluate criteria and calculate payouts
	for id, points := range counts {
		if validTargets[id] && len(points) >= f.MinCount {
			toExplode = append(toExplode, points...)

			symbolDef, exists := symbols[id]
			if exists && len(symbolDef.Payouts) > 0 {
				payoutIndex := len(points) - 1
				if payoutIndex >= len(symbolDef.Payouts) {
					payoutIndex = len(symbolDef.Payouts) - 1
				}
				totalWin += symbolDef.Payouts[payoutIndex]
			}
		}
	}

	// 4. Halt if no symbols meet the threshold
	if len(toExplode) == 0 {
		return false
	}

	// 5. Execute explosion and cascade
	explosions := grid.ExplodeAndCascade(toExplode, func(x, y int) *models.SymbolDef {
		return ctx.GetRandomSymbol(grid, x, y)
	})

	// 6. Push event to timeline
	ctx.PushTimeline(&models.TimelineEvent{
		Type:         "EXPLODE_AND_CASCADE_FEATURE",
		GridSnapshot: grid.Copy(),
		WinAmount:    totalWin,
		Meta: map[string]interface{}{
			"explosions": explosions,
		},
	})

	return true
}

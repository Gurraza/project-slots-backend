package features

import (
	"slots/internal/grid"
	"slots/internal/symbol"
	"slots/internal/timeline"
)

type PaylineFeature struct {
	BaseFeature
	Paylines         [][]int
	ExcludeSymbolIDs []int
	excludeSet       map[int]struct{}
}

func NewPaylineFeature(paylines [][]int, excludeSymbolIds []int) *PaylineFeature {
	set := make(map[int]struct{}, len(excludeSymbolIds))
	for _, id := range excludeSymbolIds {
		set[id] = struct{}{}
	}
	return &PaylineFeature{
		BaseFeature: BaseFeature{
			Type: "PAYLINES_FEATURE",
		},
		Paylines:         paylines, //GetPaylines(),
		ExcludeSymbolIDs: excludeSymbolIds,
		excludeSet:       set,
	}
}

func (f *PaylineFeature) OnGridIdle(ctx FeatureContext) bool {
	var winningLines []*LineCheckResult
	roundWin := 0.0

	for _, linePath := range f.Paylines {
		result := f.checkLine(ctx.GetGrid(), linePath, ctx.GetSymbols())
		if result != nil {
			winningLines = append(winningLines, result)
			roundWin += result.Payout
		}
	}

	if len(winningLines) > 0 {
		// eventData := map[string]interface{}{
		// 	"lines": winningLines,
		// }
		// extraData, _ := json.Marshal(eventData)
		ctx.PushTimeline(&timeline.TimelineEvent{
			Type:         f.Type,
			GridSnapshot: ctx.GetGrid().Copy(),
			WinAmount:    roundWin,
			Meta:         winningLines,
		})
	}

	return false
}

type LineCheckResult struct {
	LineID int          `json:"lineId"`
	Coords []grid.Point `json:"coords"`
	Payout float64      `json:"payout"`
	Symbol string       `json:"symbol"`
}

func (f *PaylineFeature) checkLine(g *grid.Grid, linePath []int, symbols map[int]*symbol.SymbolDef) *LineCheckResult {

	if len(linePath) != g.Cols {
		return nil
	}

	symbolInLine := make([]*symbol.SymbolDef, len(linePath))
	for col, row := range linePath {
		idHere := g.Get(col, row)
		symbHere := symbols[idHere]
		symbolInLine[col] = symbHere
	}

	var baseSymbol *symbol.SymbolDef

	for _, s := range symbolInLine {
		if !s.IsWild() {
			baseSymbol = s
			break
		}
	}

	// All wild line → treat wild as base
	if baseSymbol == nil {
		baseSymbol = symbolInLine[0]
	}

	matchCount := 1
	for i := 1; i < len(symbolInLine); i++ {
		s := symbolInLine[i]
		_, ok := f.excludeSet[s.ID]
		if baseSymbol.Compatible(s) && !ok {
			matchCount++
		} else {
			break
		}
	}
	// Add if it is a mix of different symbolDefs have payout be the lowest
	// kanske fixat redan?
	if matchCount >= 3 {
		payout := baseSymbol.Payouts[matchCount-1]
		payoutSymbol := baseSymbol

		for i := 0; i < matchCount; i++ {
			s := symbolInLine[i]
			if s.Payouts[matchCount-1] > payout {
				payout = s.Payouts[matchCount-1]
				payoutSymbol = s
			}
		}

		coords := make([]grid.Point, matchCount)
		for i := 0; i < matchCount; i++ {
			coords[i] = grid.Point{X: i, Y: linePath[i]}
		}

		return &LineCheckResult{
			Coords: coords,
			Payout: payout,
			Symbol: payoutSymbol.Name,
		}

	}

	return nil
}

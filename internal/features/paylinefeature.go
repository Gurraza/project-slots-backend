package features

import (
	"encoding/json"
	"slots/internal/grid"
	"slots/internal/symbol"
	"slots/internal/timeline"
)

type PaylineFeature struct {
	BaseFeature
	Paylines [][]int
}

func NewPaylineFeature() *PaylineFeature {
	return &PaylineFeature{
		BaseFeature: BaseFeature{
			Type: "PAYLINES_FEATURE",
		},
		Paylines: GetPaylines(),
	}
}

func (f *PaylineFeature) OnGridIdle(ctx FeatureContext) bool {
	var winningLines []interface{}
	roundWin := 0.0

	for _, linePath := range f.Paylines {
		result := f.checkLine(ctx.GetGrid(), linePath, ctx.GetSymbols())
		if result != nil {
			winningLines = append(winningLines, result)
			roundWin += result.Payout
		}
	}

	if len(winningLines) > 0 {
		eventData := map[string]interface{}{
			"lines": winningLines,
		}
		extraData, _ := json.Marshal(eventData)
		ctx.PushTimeline(&timeline.TimelineEvent{
			Type:         f.Type,
			GridSnapshot: ctx.GetGrid().Copy(),
			WinAmount:    0,
			Meta:         extraData,
		})
	}

	return false
}

func GetPaylines() [][]int {
	// Classic 5-reel, 3-row paylines
	return [][]int{
		{1, 1, 1, 1, 1}, // Line 1: middle
		{0, 0, 0, 0, 0}, // Line 2: top
		{2, 2, 2, 2, 2}, // Line 3: bottom
		{0, 1, 2, 1, 0}, // Line 4: V-shaped
		{2, 1, 0, 1, 2}, // Line 5: inverted V
	}
}

type LineCheckResult struct {
	LineID int          `json:"lineId"`
	Coords []grid.Point `json:"coords"`
	Payout float64      `json:"payout"`
	Symbol string       `json:"symbol"`
}

func (f *PaylineFeature) checkLine(g *grid.Grid, linePath []int, allSymbols []*symbol.SymbolDef) *LineCheckResult {
	if len(linePath) != len(g.Cells) {
		return nil
	}

	symbols := make([]int, len(linePath))
	for col, row := range linePath {
		if col >= len(g.Cells) || row >= len(g.Cells[col]) {
			return nil
		}
		symbols[col] = g.Cells[col][row]
	}

	firstSymbol := symbols[0]
	if firstSymbol == 0 {
		return nil
	}

	matchCount := 1
	for i := 1; i < len(symbols); i++ {
		if symbols[i] == firstSymbol {
			matchCount++
		} else {
			break
		}
	}

	if matchCount >= 3 {
		symbolDef := allSymbols[firstSymbol]
		payout := symbolDef.Payouts[matchCount-1]

		coords := make([]grid.Point, matchCount)
		for i := 0; i < matchCount; i++ {
			coords[i] = grid.Point{X: i, Y: linePath[i]}
		}

		return &LineCheckResult{
			Coords: coords,
			Payout: payout,
			Symbol: symbolDef.Name,
		}

	}

	return nil
}

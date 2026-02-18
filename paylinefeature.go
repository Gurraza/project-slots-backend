package main

import (
	"encoding/json"
)

type PaylineFeature struct {
	BaseFeature
	Paylines [][]int
}

var _ GameFeature = (*PaylineFeature)(nil)

func NewPaylineFeature(roundSession *GameState) *PaylineFeature {
	return &PaylineFeature{
		BaseFeature: BaseFeature{
			Type:      "PAYLINES_FEATURE",
			GameState: roundSession,
		},
		Paylines: GetPaylines(),
	}
}

func (f *PaylineFeature) OnGridIdle(session *GameState) bool {
	var winningLines []interface{}
	roundWin := 0.0

	for _, linePath := range f.Paylines {
		result := f.checkLine(*session.Grid, linePath)
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

		session.Timeline = append(session.Timeline, TimelineEvent{
			Type:         f.Type,
			GridSnapshot: session.Grid.Copy(),
			WinAmount:    0,
			Meta:         extraData,
		})
		return false
	}

	return false
}

type LineCheckResult struct {
	LineID int     `json:"lineId"`
	Coords []Point `json:"coords"`
	Payout float64 `json:"payout"`
	Symbol string  `json:"symbol"`
}

func (f *PaylineFeature) checkLine(grid Grid, linePath []int) *LineCheckResult {
	if len(linePath) != len(grid.Cells) {
		return nil
	}

	symbols := make([]int, len(linePath))
	for col, row := range linePath {
		if col >= len(grid.Cells) || row >= len(grid.Cells[col]) {
			return nil
		}
		symbols[col] = grid.Cells[col][row]
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
		symbolDef := f.GameState.Config.Symbols[firstSymbol]
		payout := symbolDef.Payouts[matchCount]

		coords := make([]Point, matchCount)
		for i := 0; i < matchCount; i++ {
			coords[i] = Point{X: i, Y: linePath[i]}
		}

		return &LineCheckResult{
			Coords: coords,
			Payout: payout,
			Symbol: symbolDef.Name,
		}

	}

	return nil
}

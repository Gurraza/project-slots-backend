package main

func LoadLinesConfig() *GameConfig {
	return &GameConfig{
		Cols: 5,
		Rows: 3,
		Symbols: []SymbolDef{
			{ID: 1, Name: "wild", Weight: []int{20}, Payouts: []float64{0, 0, 5, 50, 1000}},
			{ID: 3, Name: "bar1", Weight: []int{28}, Payouts: []float64{0, 0, 3, 25, 200}},
			{ID: 4, Name: "bar2", Weight: []int{28}, Payouts: []float64{0, 0, 3, 25, 200}},
			{ID: 5, Name: "bar3", Weight: []int{28}, Payouts: []float64{0, 0, 3, 25, 200}},
			{ID: 0, Name: "mixed_bar", Weight: []int{0}, MatchesWith: []string{""}},
			{ID: 2, Name: "strawberry", Weight: []int{15}, Payouts: []float64{0, 0, 5, 50, 500}},
			{ID: 6, Name: "cherry", Weight: []int{60}, Payouts: []float64{0, 0, 1, 8, 40}},
		},
	}
}

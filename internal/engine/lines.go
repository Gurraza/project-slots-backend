package engine

// func LoadLinesConfig() *GameConfig {
// 	return &GameConfig{
// 		Cols: 5,
// 		Rows: 3,
// 		Symbols: []SymbolDef{
// 			// {ID: 3, Name: "bar1", Weight: []int{28}, Payouts: []float64{0, 0, 3, 25, 200}},
// 			// {ID: 4, Name: "bar2", Weight: []int{28}, Payouts: []float64{0, 0, 3, 25, 200}},
// 			// {ID: 0, Name: "mixed_bar", Weight: []int{0}, MatchesWith: []string{""}},
// 			// {ID: 5, Name: "bar3", Weight: CompositeWeight{
// 			// 	Providers: []WeightProvider{
// 			// 		FixedWeight{Value: 5},
// 			// 	},
// 			// }, Payouts: []float64{0, 0, 3, 25, 200}},
// 			// {ID: 2, Name: "strawberry", Weight: FixedWeight{Value: 10}, Payouts: []float64{0, 0, 5, 50, 500}},
// 			{ID: 0, Name: "cherry", WeightConfig: WeightConfig{FixedWeight: 100, Modifiers: []WeightModifier{}}, Payouts: []float64{0, 0, 1, 8, 40}},
// 			{ID: 1, Name: "orange", WeightConfig: WeightConfig{FixedWeight: 100, Modifiers: []WeightModifier{}}, Payouts: []float64{0, 0, 1, 8, 40}},
// 			{ID: 2, Name: "citrus", WeightConfig: WeightConfig{FixedWeight: 100, Modifiers: []WeightModifier{}}, Payouts: []float64{0, 0, 1, 8, 40}},
// 			{ID: 3, Name: "asdfgh", WeightConfig: WeightConfig{FixedWeight: 100, Modifiers: []WeightModifier{
// 				&ReelWeight{[]int{0, 0, 0, 100, 100}},
// 				&CountWeight{[]int{100, 100, 0, 0, 0, 0}},
// 			}}, Payouts: []float64{0, 0, 1, 8, 40}},
// 		},
// 		Features: []GameFeature{
// 			// NewWildFeature(SymbolDef{ID: 11, Name: "wild", Weight: CompositeWeight{
// 			// 	Providers: []WeightProvider{
// 			// 		CountWeight{[]int{15, 15, 15, 10, 5, 2}},
// 			// 		ReelWeight{.5, .6, .7, .8, .9},
// 			// 	},
// 			// }, Payouts: []float64{0, 0, 5, 50, 1000}}),
// 			NewPaylineFeature(),
// 		},
// 	}
// }

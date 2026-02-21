package engine

import (
	"slots/internal/features"
	"slots/internal/grid"
	"slots/internal/rng"
	"slots/internal/symbol"
	"slots/internal/timeline"
)

type GameConfig struct {
	Cols     int
	Rows     int
	Symbols  []*symbol.SymbolDef
	Features []features.GameFeature
}

type GameState struct {
	Grid           *grid.Grid
	Config         *GameConfig
	Symbols        map[int]*symbol.SymbolDef
	RNG            rng.RNG
	Timeline       []*timeline.TimelineEvent
	ActiveFeatures []features.GameFeature
}

func (g *GameState) GetGrid() *grid.Grid {
	return g.Grid
}

func (g *GameState) AddSymbol(s *symbol.SymbolDef) {
	if g.Symbols[s.ID] != nil {
		panic("Tried to add a symbol with id that already exists")
	}
	g.Symbols[s.ID] = s
}

func (g *GameState) GetSymbols() map[int]*symbol.SymbolDef {
	return g.Symbols
}

func (gs *GameState) GetRandomSymbol(grid *grid.Grid, col int, row int) *symbol.SymbolDef {
	weightContext := symbol.WeightContext{
		ReelIndex: col,
		RowIndex:  row,
		Grid:      grid,
	}
	totalWeight := 0
	for _, s := range gs.Symbols {
		totalWeight += s.GetWeight(&weightContext)
	}
	randomNmr := gs.RNG.Range(totalWeight)

	for _, s := range gs.Symbols {
		w := s.GetWeight(&weightContext)
		if randomNmr < w {
			return s
		}
		randomNmr -= w
	}
	panic("GET RANDOM SYMBOL DIDN't RETURN ANYTHING CORRECT")
	// will never happen
	// return gs.Symbols[0]
}

func (g *GameState) PushTimeline(e *timeline.TimelineEvent) {
	if len(g.Timeline) == 0 {
		e.TotalWinAmount = 0
	} else {
		e.TotalWinAmount = g.Timeline[len(g.Timeline)-1].TotalWinAmount + e.WinAmount
	}
	g.Timeline = append(g.Timeline, e)
}

func (g *GameState) SetGrid(newGrid *grid.Grid) {
	g.Grid = newGrid
}

func NewGameState(config *GameConfig, seed int64) *GameState {
	symbs := make(map[int]*symbol.SymbolDef, len(config.Symbols))
	for _, s := range config.Symbols {
		symbs[s.ID] = s
	}

	activeFeatures := make([]features.GameFeature, len(config.Features))
	copy(activeFeatures, config.Features)

	return &GameState{
		Grid:           grid.NewGrid(config.Cols, config.Rows),
		Config:         config,
		RNG:            rng.NewGoRNG(seed), // Initialize with the specific seed
		Symbols:        symbs,
		Timeline:       make([]*timeline.TimelineEvent, 0),
		ActiveFeatures: activeFeatures,
	}
}

func (g *GameState) AddFeature(f features.GameFeature) {
	g.ActiveFeatures = append(g.ActiveFeatures, f)
}

// func (g *GameState) RemoveFeature(featureID string) {
// 	for i, feature := range g.ActiveFeatures {
// 		// Assumes GameFeature interface has an ID() string method
// 		if feature.GetType() == featureID {
// 			// Order-preserving removal
// 			copy(g.ActiveFeatures[i:], g.ActiveFeatures[i+1:])
// 			g.ActiveFeatures[len(g.ActiveFeatures)-1] = nil // Prevent memory leak
// 			g.ActiveFeatures = g.ActiveFeatures[:len(g.ActiveFeatures)-1]
// 			return
// 		}
// 	}
// }

func (g *GameState) RemoveFeature(featureID string) {
	var keptFeatures []features.GameFeature

	for _, feature := range g.ActiveFeatures {
		if feature.GetType() == featureID {
			// 1. Fetch symbols owned by the feature being removed
			symbolsToRemove := feature.GetSymbols(g)

			for _, symToRemove := range symbolsToRemove {
				// 2. Safety check: Do not delete if it is a base game symbol
				isBaseSymbol := false
				for _, baseSym := range g.Config.Symbols {
					if baseSym.ID == symToRemove.ID {
						isBaseSymbol = true
						break
					}
				}

				// 3. Remove from GameState map if it belongs exclusively to the feature
				if !isBaseSymbol {
					delete(g.Symbols, symToRemove.ID)
				}
			}
			// Do not append this feature to keptFeatures
		} else {
			keptFeatures = append(keptFeatures, feature)
		}
	}

	// Replace the old slice with the filtered one
	g.ActiveFeatures = keptFeatures
}

func (config *GameConfig) PlayGame(seed int64) []*timeline.TimelineEvent {
	gameState := NewGameState(config, seed)
	for _, f := range gameState.ActiveFeatures {
		for _, s := range f.GetSymbols(gameState) {
			gameState.AddSymbol(s)
		}
		f.Init(gameState)
	}

	gameState.PushTimeline(&timeline.TimelineEvent{
		Type:         "SPIN_START",
		GridSnapshot: gameState.Grid.Copy(),
		WinAmount:    0,
	})
	t := gameState.Spin()

	return t
}

func (gameState *GameState) Spin() []*timeline.TimelineEvent {
	gameState.Grid = GenerateRandomGrid(gameState)

	for _, f := range gameState.ActiveFeatures {
		f.OnSpinStart(gameState)
	}

	count := 0
	for {
		actionOccured := false
		if count > 50 {
			break
		}

		for _, f := range gameState.ActiveFeatures {
			if f.OnGridEvaluate(gameState) {
				actionOccured = true
			}
		}

		if !actionOccured {
			for _, f := range gameState.ActiveFeatures {
				if f.OnGridIdle(gameState) {
					actionOccured = true
				}
			}
		}

		count += 1
		if !actionOccured {
			break
		}
	}

	for _, f := range gameState.ActiveFeatures {
		f.OnSpinEnd(gameState)
	}

	// gameState.PushTimeline(&timeline.TimelineEvent{
	// 	Type:         "GAME_OVER",
	// 	GridSnapshot: gameState.Grid.Copy(),
	// 	WinAmount:    0,
	// })
	return gameState.Timeline
}

func GenerateRandomGrid(gameState *GameState) *grid.Grid {
	totalCells := gameState.Grid.Cols * gameState.Grid.Rows
	newGrid := grid.NewGrid(gameState.Grid.Cols, gameState.Grid.Rows)

	indices := make([]int, totalCells)
	for i := range indices {
		indices[i] = i
	}

	// Fisher-Yates Shuffle
	for i := totalCells - 1; i > 0; i-- {
		j := gameState.RNG.Intn(i + 1)
		indices[i], indices[j] = indices[j], indices[i]
	}

	for _, idx := range indices {
		col := idx / gameState.Grid.Rows
		row := idx % gameState.Grid.Rows

		newGrid.Cells[col][row] = gameState.GetRandomSymbol(newGrid, col, row).ID
	}

	return newGrid
}

func (g *GameState) GetSymbol(id int) *symbol.SymbolDef {
	return g.Symbols[id] // Returns nil if not found, which is standard

	// for _, s := range g.Symbols {
	// 	if s.ID == id {
	// 		return s
	// 	}
	// }
	// return nil
}

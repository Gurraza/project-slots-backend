package engine

import (
	"slots/internal/features"
	"slots/internal/models"
	"sort"
)

type GameConfig struct {
	Cols     int
	Rows     int
	Symbols  []*models.SymbolDef
	Features []features.GameFeature
}

type GameState struct {
	Grid           *models.Grid
	Config         *GameConfig
	Symbols        map[int]*models.SymbolDef
	RNG            models.RNG
	Timeline       []*models.TimelineEvent
	ActiveFeatures []features.GameFeature
}

func (g *GameState) GetGrid() *models.Grid {
	return g.Grid
}

func (g *GameState) AddSymbol(s *models.SymbolDef) {
	if g.Symbols[s.ID] != nil {
		panic("Tried to add a symbol with id that already exists")
	}
	g.Symbols[s.ID] = s
}

func (g *GameState) GetSymbols() map[int]*models.SymbolDef {
	return g.Symbols
}

func (gs *GameState) GetRandomSymbol(grid *models.Grid, col int, row int) *models.SymbolDef {
	keys := make([]int, 0, len(gs.Symbols))
	for k := range gs.Symbols {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	weightContext := models.WeightContext{
		ReelIndex: col,
		RowIndex:  row,
		Grid:      grid,
	}
	totalWeight := 0
	for _, k := range keys {
		totalWeight += gs.Symbols[k].GetWeight(&weightContext)
	}
	randomNmr := gs.RNG.Range(totalWeight)

	for _, k := range keys {
		s := gs.Symbols[k]
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

func (g *GameState) PushTimeline(e *models.TimelineEvent) {
	if len(g.Timeline) == 0 {
		e.TotalWinAmount = 0
	} else {
		e.TotalWinAmount = g.Timeline[len(g.Timeline)-1].TotalWinAmount + e.WinAmount
	}
	g.Timeline = append(g.Timeline, e)
}

func (g *GameState) SetGrid(newGrid *models.Grid) {
	g.Grid = newGrid
}

func NewGameState(config *GameConfig, seed int64) *GameState {
	symbs := make(map[int]*models.SymbolDef, len(config.Symbols))
	for _, s := range config.Symbols {
		symbs[s.ID] = s
	}

	activeFeatures := make([]features.GameFeature, len(config.Features))
	copy(activeFeatures, config.Features)

	return &GameState{
		Grid:           models.NewGrid(config.Cols, config.Rows),
		Config:         config,
		RNG:            models.NewGoRNG(seed), // Initialize with the specific seed
		Symbols:        symbs,
		Timeline:       make([]*models.TimelineEvent, 0),
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

func (config *GameConfig) PlayGame(seed int64) []*models.TimelineEvent {
	gameState := NewGameState(config, seed)
	for _, f := range gameState.ActiveFeatures {
		for _, s := range f.GetSymbols(gameState) {
			gameState.AddSymbol(s)
		}
		f.Init(gameState)
	}

	t := gameState.Spin()

	return t
}

func (gameState *GameState) Spin() []*models.TimelineEvent {
	gameState.Grid = GenerateRandomGrid(gameState)

	gameState.PushTimeline(&models.TimelineEvent{
		Type:         "SPIN_START",
		GridSnapshot: gameState.Grid.Copy(),
		WinAmount:    0,
		Meta:         gameState.RNG.GetSeed(),
	})

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

func GenerateRandomGrid(gameState *GameState) *models.Grid {
	totalCells := gameState.Grid.Cols * gameState.Grid.Rows
	newGrid := models.NewGrid(gameState.Grid.Cols, gameState.Grid.Rows)

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

func (g *GameState) GetSymbol(id int) *models.SymbolDef {
	return g.Symbols[id]
}

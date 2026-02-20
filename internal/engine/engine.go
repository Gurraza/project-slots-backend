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
	Grid     *grid.Grid
	Config   *GameConfig
	Symbols  map[int]*symbol.SymbolDef
	RNG      rng.RNG
	Timeline []*timeline.TimelineEvent
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
	return &GameState{
		Grid:     grid.NewGrid(config.Cols, config.Rows),
		Config:   config,
		RNG:      rng.NewGoRNG(seed), // Initialize with the specific seed
		Symbols:  symbs,
		Timeline: make([]*timeline.TimelineEvent, 0),
	}
}

func (gameState *GameState) PlayRound() []*timeline.TimelineEvent {
	for _, f := range gameState.Config.Features {
		for _, s := range f.GetSymbols(gameState) {
			gameState.AddSymbol(s)
		}
		f.Init(gameState)
	}

	gameState.Grid = GenerateRandomGrid(gameState)

	gameState.PushTimeline(&timeline.TimelineEvent{
		Type:         "SPIN_START",
		GridSnapshot: gameState.Grid.Copy(),
		WinAmount:    0,
	})

	for _, f := range gameState.Config.Features {
		f.OnSpinStart(gameState)
	}

	count := 0
	for {
		actionOccured := false
		if count > 50 {
			break
		}

		for _, f := range gameState.Config.Features {
			actionOccured = f.OnGridEvaluate(gameState)
		}

		if !actionOccured {
			for _, f := range gameState.Config.Features {
				actionOccured = f.OnGridIdle(gameState)
			}
		}

		count += 1
		if !actionOccured {
			break
		}
	}

	for _, f := range gameState.Config.Features {
		f.OnSpinEnd(gameState)
	}

	gameState.PushTimeline(&timeline.TimelineEvent{
		Type:         "GAME_OVER",
		GridSnapshot: gameState.Grid.Copy(),
		WinAmount:    0,
	})
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
	for _, s := range g.Symbols {
		if s.ID == id {
			return s
		}
	}
	return nil
}

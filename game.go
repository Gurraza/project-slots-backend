package main

type GameConfig struct {
	Cols     int
	Rows     int
	Symbols  []SymbolDef
	Features []GameFeature
}

type SymbolDef struct {
	ID          int
	Name        string
	Weight      []int
	MatchesWith []string
	Payouts     []float64
}

func (s *SymbolDef) GetWeight(gameState *GameState) int {
	count := len(gameState.Grid.Contain(s.ID))
	if count >= len(s.Weight) {
		return s.Weight[len(s.Weight)-1] // return the last weight
	}
	return s.Weight[count]
}

type TimelineEvent struct {
	Type         string      `json:"type"`
	GridSnapshot *Grid       `json:"grid"`
	WinAmount    float64     `json:"win"`
	Meta         interface{} `json:"meta,omitempty"`
}

type GameState struct {
	Grid     *Grid
	Config   *GameConfig
	RNG      RNG
	Timeline []TimelineEvent
}

func NewGameState(config *GameConfig, seed int64) *GameState {
	return &GameState{
		Grid:     NewGrid(config.Cols, config.Rows),
		Config:   config,
		RNG:      NewGoRNG(seed), // Initialize with the specific seed
		Timeline: make([]TimelineEvent, 0),
	}
}

func PlayRound(gameState *GameState) []TimelineEvent {
	gameState.Grid = gameState.Grid.GenerateRandomGrid(gameState)
	gameState.Timeline = append(gameState.Timeline, TimelineEvent{
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

	gameState.Timeline = append(gameState.Timeline, TimelineEvent{
		Type:         "GAME_OVER",
		GridSnapshot: gameState.Grid.Copy(),
		WinAmount:    0,
	})
	return gameState.Timeline
}

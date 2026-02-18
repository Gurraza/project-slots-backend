package main

type GameConfig struct {
	Cols    int
	Rows    int
	Symbols []SymbolDef
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
	Features []GameFeature
}

func NewGameState(config *GameConfig, seed int64) *GameState {
	return &GameState{
		Grid:     NewGrid(config.Cols, config.Rows),
		Config:   config,
		RNG:      NewGoRNG(seed), // Initialize with the specific seed
		Timeline: make([]TimelineEvent, 0),
		Features: []GameFeature{},
	}
}

func PlayRound(gameState *GameState) []TimelineEvent {
	gameState.Grid = gameState.Grid.GenerateRandomGrid(gameState)
	gameState.Timeline = append(gameState.Timeline, TimelineEvent{
		Type:         "SPIN_START",
		GridSnapshot: gameState.Grid.Copy(),
		WinAmount:    0,
	})

	for i := range gameState.Features {
		gameState.Features[i].OnSpinStart(gameState)
	}
	count := 0
	for {
		if count > 20 {
			break
		}

		cluster := FindClusters(*gameState.Grid, gameState.Config.Symbols)
		mergedCluster := make([]Point, 0)
		for _, c := range cluster {
			mergedCluster = append(mergedCluster, c...)
		}
		if len(mergedCluster) == 0 {
			break
		}

		replacements := make([]int, len(mergedCluster))
		for i := range replacements {
			replacements[i] = gameState.RNG.GetRandomSymbol(gameState).ID
		}
		gameState.Grid, _ = ExplodeAndCascade(*gameState.Grid, mergedCluster, replacements)

		gameState.Timeline = append(gameState.Timeline, TimelineEvent{
			Type:         "ExplodeAndCascade",
			GridSnapshot: gameState.Grid.Copy(),
			WinAmount:    0,
			Meta: map[string]interface{}{
				"points":       mergedCluster,
				"replacements": replacements,
			},
		})
		count += 1
	}

	gameState.Timeline = append(gameState.Timeline, TimelineEvent{
		Type:         "GAME_OVER",
		GridSnapshot: gameState.Grid.Copy(),
		WinAmount:    0,
	})
	return gameState.Timeline
}

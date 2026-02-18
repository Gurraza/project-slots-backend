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

func PlayRound(roundSession *GameState) []TimelineEvent {
	roundSession.Timeline = append(roundSession.Timeline, TimelineEvent{
		Type:         "SPIN_START",
		GridSnapshot: roundSession.Grid.Copy(),
		WinAmount:    0,
	})

	for i := range roundSession.Features {
		roundSession.Features[i].OnSpinStart(roundSession)
	}

	return roundSession.Timeline
}

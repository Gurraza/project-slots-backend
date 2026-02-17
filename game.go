package main

type GameConfig struct {
	Symbols []SymbolDef
}

type SymbolDef struct {
	ID          int
	Name        string
	Weight      []int
	MatchesWith []string
}

type EventType int

const (
	EventSpinStart EventType = iota
	EventExplode
	EventInsert
	EventSpinEnd
)

// TimelineEvent records what happened
type TimelineEvent struct {
	Type         EventType   `json:"type"`
	GridSnapshot *Grid       `json:"grid"`
	WinAmount    float64     `json:"win"`
	Meta         interface{} `json:"meta,omitempty"` // For clusters, replacements, etc
}

type RoundSession struct {
	Grid        *Grid
	Config      *GameConfig
	RNG         RNG
	Timeline    []TimelineEvent
	BaseFeature []BaseFeature
}

func NewRound(config *GameConfig, seed int64) *RoundSession {
	return &RoundSession{
		Config:   config,
		RNG:      NewGoRNG(seed), // Initialize with the specific seed
		Timeline: make([]TimelineEvent, 0),
	}
}

func playRound(roundSession *RoundSession) {
	roundSession.Timeline = append(roundSession.Timeline, TimelineEvent{
		Type:         EventSpinStart,
		GridSnapshot: roundSession.Grid.Copy(),
		WinAmount:    0,
	})
}

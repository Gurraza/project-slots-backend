package models

type TimelineEvent struct {
	Type           string  `json:"type"`
	GridSnapshot   *Grid   `json:"grid"`
	WinAmount      float64 `json:"win"`
	TotalWinAmount float64 `json:"totalWin"`
	Meta           any     `json:"meta"`
}

type Timeline []*TimelineEvent

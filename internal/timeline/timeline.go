package timeline

import "slots/internal/grid"

type TimelineEvent struct {
	Type           string     `json:"type"`
	GridSnapshot   *grid.Grid `json:"grid"`
	WinAmount      float64    `json:"win"`
	TotalWinAmount float64    `json:"to_twin"`
	Meta           any        `json:"meta"`
}

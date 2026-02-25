package timeline

import (
	"fmt"
	"slots/internal/grid"
)

type TimelineEvent struct {
	Type           string     `json:"type"`
	GridSnapshot   *grid.Grid `json:"grid"`
	WinAmount      float64    `json:"win"`
	TotalWinAmount float64    `json:"to_twin"`
	Meta           any        `json:"meta"`
}

type Timeline []*TimelineEvent

func (t Timeline) Print() {

	fmt.Printf("Events: %d Total Win Amount %.2f \n", len(t), t[len(t)-1].TotalWinAmount)
	for i, e := range t {
		fmt.Printf("Grid %d type "+e.Type+" win amount %.2f\n", i, e.WinAmount)

		// if e.Meta != nil {
		// 	// MarshalIndent formats the JSON with newlines and spaces
		// 	metaBytes, err := json.MarshalIndent(e.Meta, "", "  ")
		// 	if err != nil {
		// 		fmt.Printf("Meta JSON Error: %v\n", err)
		// 	} else {
		// 		// Convert the byte slice to a string to print characters instead of numbers
		// 		fmt.Println(string(metaBytes))
		// 	}
		// }

		g := e.GridSnapshot.Cells
		if len(g) == 0 {
			continue
		}

		cols := len(g)
		rows := len(g[0])

		// 1. Loop through rows (Y axis) first
		for y := 0; y < rows; y++ {
			// 2. Loop through columns (X axis) for the current row
			for x := 0; x < cols; x++ {
				// Use %3d to pad single-digit numbers so columns align perfectly
				fmt.Printf("%3d ", g[x][y])
			}
			// 3. Print a newline at the end of each row
			fmt.Println()
		}
		fmt.Println()
	}
}

package models

import (
	"encoding/json"
)

type Grid struct {
	Rows  int     `json:"rows"`
	Cols  int     `json:"cols"`
	Cells [][]int `json:"cells"`
}

type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

func (g *Grid) Get(x int, y int) int {
	if x >= 0 && x < g.Cols && y >= 0 && y < g.Rows {
		return g.Cells[x][y]
	} else {
		panic("Called Grid.Get with parameters that are outside of the grid")
	}
}

func (g *Grid) Set(x int, y int, value int) int {
	if x >= 0 && x < g.Cols && y >= 0 && y < g.Rows {
		before := g.Cells[x][y]
		g.Cells[x][y] = value
		return before
	} else {
		panic("Called Grid.Set with parameters that are outside of the grid")
	}
}

func (g *Grid) MarshalJSON() ([]byte, error) {
	// We only want the [][]int field to be sent to the client
	return json.Marshal(g.Cells)
}

func NewGrid(cols, rows int) *Grid {
	var cells [][]int = make([][]int, cols)
	for i := range cells {
		cells[i] = make([]int, rows)
	}
	return &Grid{Cols: cols, Rows: rows, Cells: cells}
}

func (g *Grid) Copy() *Grid {
	newCells := make([][]int, g.Cols)
	for i := range g.Cells {
		newCells[i] = make([]int, g.Rows)
		copy(newCells[i], g.Cells[i])
	}
	return &Grid{Cols: g.Cols, Rows: g.Rows, Cells: newCells}
}

// ExplodeAndCascade removes points, collapses columns (gravity), and refills from the top.
func ExplodeAndCascade(inputGrid *Grid, points []Point, replacements []int) *Grid {
	// 1. Create a Deep Copy of the Grid to ensure immutability
	newGrid := inputGrid.Copy()

	// 2. Validation: Ensure we have enough replacements
	// Note: If you allow clusters of same-value-but-different-points (overlapping),
	// deduping points might be necessary here to get the exact count.
	if len(replacements) < len(points) {
		panic("not enough replacement integers provided for the number of exploded points")
	}

	// 3. Mark points for removal
	// Using -1 to represent a "hole" temporarily.
	// We use a map for O(1) lookups if the points slice is large,
	// but for small clusters, direct iteration is fine.
	// Here we assume standard removal logic.
	for _, p := range points {
		if p.X >= 0 && p.X < newGrid.Cols && p.Y >= 0 && p.Y < newGrid.Rows {
			newGrid.Cells[p.X][p.Y] = -1
		}
	}

	replacementIndex := 0

	// 4. Process each column to apply Gravity and Replenish
	for x := 0; x < newGrid.Cols; x++ {
		// Step A: Collect "surviving" blocks (non -1) in this column
		survivors := []int{}
		for y := 0; y < newGrid.Rows; y++ {
			val := newGrid.Cells[x][y]
			if val != -1 {
				survivors = append(survivors, val)
			}
		}

		// Step B: Calculate how many new blocks are needed at the top
		missingCount := newGrid.Rows - len(survivors)

		// Step C: Rebuild the column
		// 1. Add new replacements at the top
		for i := 0; i < missingCount; i++ {
			newGrid.Cells[x][i] = replacements[replacementIndex]
			replacementIndex++
		}

		// 2. Append the surviving blocks below the new ones
		for i := 0; i < len(survivors); i++ {
			newGrid.Cells[x][missingCount+i] = survivors[i]
		}
	}

	return newGrid
}

func (g *Grid) Contain(id int) []Point {
	positions := []Point{}
	for i := range g.Cols {
		for j := range g.Rows {
			if g.Cells[i][j] == id {
				positions = append(positions, Point{X: i, Y: j})
			}
		}
	}
	return positions
}

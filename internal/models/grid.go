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

type ExplosionPoint struct {
	X             int `json:"x"`
	Y             int `json:"y"`
	ReplacementID int `json:"replacementId"`
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

func (g *Grid) ExplodeAndCascade(points []Point, getRandomSymbol func(x, y int) *SymbolDef) []*ExplosionPoint {
	explosions := make([]*ExplosionPoint, 0)
	// Group replacements by column
	replacementsByCol := make(map[int][]int)
	for _, exp := range points {
		if exp.X >= 0 && exp.X < g.Cols && exp.Y >= 0 && exp.Y < g.Rows {
			g.Cells[exp.X][exp.Y] = -1
			newSym := getRandomSymbol(exp.X, exp.Y)

			explosions = append(explosions, &ExplosionPoint{
				X:             exp.X,
				Y:             exp.Y,
				ReplacementID: newSym.ID,
			})

			replacementsByCol[exp.X] = append(replacementsByCol[exp.X], newSym.ID)
		}
	}

	// For each column, apply gravity and replenish
	for x := range g.Cols {
		survivors := []int{}
		for y := 0; y < g.Rows; y++ {
			val := g.Cells[x][y]
			if val != -1 {
				survivors = append(survivors, val)
			}
		}

		missingCount := g.Rows - len(survivors)
		colReplacements := replacementsByCol[x]

		if len(colReplacements) < missingCount {
			panic("not enough replacement integers provided for column")
		}

		// Insert replacements from the top
		for i := 0; i < missingCount; i++ {
			g.Cells[x][i] = colReplacements[i]
		}

		// Insert survivors
		for i := 0; i < len(survivors); i++ {
			g.Cells[x][missingCount+i] = survivors[i]
		}
	}

	return explosions
}

// ExplodeAndCascade removes points, collapses columns (gravity), and refills from the top.
func eExplodeAndCascade(inputGrid *Grid, points []Point, replacements []int) *Grid {
	newGrid := inputGrid.Copy()

	if len(replacements) < len(points) {
		panic("not enough replacement integers provided for the number of exploded points")
	}

	// Check valid coordinate in grid
	for _, p := range points {
		if p.X >= 0 && p.X < newGrid.Cols && p.Y >= 0 && p.Y < newGrid.Rows {
			newGrid.Cells[p.X][p.Y] = -1
		}
	}

	replacementIndex := 0

	// For each column, apply gravity and replenish
	for x := range newGrid.Cols {
		// Calculate amount of surviving ids
		survivors := []int{}
		for y := 0; y < newGrid.Rows; y++ {
			val := newGrid.Cells[x][y]
			if val != -1 {
				survivors = append(survivors, val)
			}
		}

		// Amounts of -1
		missingCount := newGrid.Rows - len(survivors)

		// Start by inserting the replacements
		for i := range missingCount {
			newGrid.Cells[x][i] = replacements[replacementIndex]
			replacementIndex++
		}

		// Then we insert the survivors
		for i := 0; i < len(survivors); i++ {
			newGrid.Cells[x][missingCount+i] = survivors[i]
		}
	}

	return newGrid
}

func (g *Grid) Contain(id int) []*Point {
	positions := []*Point{}
	for i := range g.Cols {
		for j := range g.Rows {
			if g.Cells[i][j] == id {
				positions = append(positions, &Point{X: i, Y: j})
			}
		}
	}
	return positions
}

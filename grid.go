package main

type Grid struct {
	Rows  int
	Cols  int
	Cells [][]int
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

func GetPaylines() [][]int {
	// Classic 5-reel, 3-row paylines
	return [][]int{
		{1, 1, 1, 1, 1}, // Line 1: middle
		{0, 0, 0, 0, 0}, // Line 2: top
		{2, 2, 2, 2, 2}, // Line 3: bottom
		{0, 1, 2, 1, 0}, // Line 4: V-shaped
		{2, 1, 0, 1, 2}, // Line 5: inverted V
	}
}

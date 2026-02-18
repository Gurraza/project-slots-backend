package main

import (
	"encoding/json"
	"errors"
)

type Grid struct {
	Rows  int     `json:"rows"`
	Cols  int     `json:"cols"`
	Cells [][]int `json:"cells"`
}

func (g Grid) MarshalJSON() ([]byte, error) {
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

func (g *Grid) GenerateRandomGrid(gameState *GameState) *Grid {
	newCells := make([][]int, g.Cols)
	for i := range g.Cells {
		newCells[i] = make([]int, g.Rows)
		for j := range g.Rows {
			newCells[i][j] = gameState.RNG.Range(len(gameState.Config.Symbols))
		}
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

func FindClusters(grid Grid, defs []SymbolDef) [][]Point {
	// 1. Build a lookup map for efficient SymbolDef access by ID
	defMap := make(map[int]SymbolDef)
	for _, d := range defs {
		defMap[d.ID] = d
	}

	// 2. Initialize visited matrix to keep track of processed cells
	visited := make([][]bool, grid.Cols)
	for x := 0; x < grid.Cols; x++ {
		visited[x] = make([]bool, grid.Rows)
	}

	var clusters [][]Point

	// Helper function to check if two IDs are compatible
	areCompatible := func(id1, id2 int) bool {
		// Exact match is always a connection
		if id1 == id2 {
			return true
		}

		def1, exists1 := defMap[id1]
		def2, exists2 := defMap[id2]

		// If either definition is missing, we rely solely on exact ID match (checked above)
		if !exists1 || !exists2 {
			return false
		}

		// Check if def1 matches def2 (by specific name or wildcard)
		for _, matchName := range def1.MatchesWith {
			if matchName == "*" || matchName == def2.Name {
				return true
			}
		}

		// Check if def2 matches def1 (symmetric check)
		for _, matchName := range def2.MatchesWith {
			if matchName == "*" || matchName == def1.Name {
				return true
			}
		}

		return false
	}

	// 3. Iterate through every cell in the grid
	for x := 0; x < grid.Cols; x++ {
		for y := 0; y < grid.Rows; y++ {
			// If already visited, skip
			if visited[x][y] {
				continue
			}

			// Start a new cluster
			currentCluster := []Point{}

			// Queue for BFS
			queue := []Point{{X: x, Y: y}}
			visited[x][y] = true

			// Perform Flood Fill (BFS)
			for len(queue) > 0 {
				curr := queue[0]
				queue = queue[1:] // Dequeue

				// Add current point to the cluster
				currentCluster = append(currentCluster, curr)

				// Define 4-way directions (Up, Down, Left, Right)
				dirs := []Point{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}

				for _, d := range dirs {
					nx, ny := curr.X+d.X, curr.Y+d.Y

					// Boundary Check
					if nx >= 0 && nx < grid.Cols && ny >= 0 && ny < grid.Rows {
						if !visited[nx][ny] {
							currentVal := grid.Cells[curr.X][curr.Y]
							neighborVal := grid.Cells[nx][ny]

							// Compatibility Check
							if areCompatible(currentVal, neighborVal) {
								visited[nx][ny] = true
								queue = append(queue, Point{X: nx, Y: ny})
							}
						}
					}
				}
			}

			// Only add non-empty clusters (though by logic size is always >= 1)
			if len(currentCluster) > 1 {
				clusters = append(clusters, currentCluster)
			}
		}
	}

	return clusters
}

// ExplodeAndCascade removes points, collapses columns (gravity), and refills from the top.
func ExplodeAndCascade(inputGrid Grid, points []Point, replacements []int) (*Grid, error) {
	// 1. Create a Deep Copy of the Grid to ensure immutability
	newGrid := inputGrid.Copy()

	// 2. Validation: Ensure we have enough replacements
	// Note: If you allow clusters of same-value-but-different-points (overlapping),
	// deduping points might be necessary here to get the exact count.
	if len(replacements) < len(points) {
		return &inputGrid, errors.New("not enough replacement integers provided for the number of exploded points")
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

	return newGrid, nil
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

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

// Takes a grid and symbolmap where id points to its definitoin and a clusterSize.
// Returns a slice of clusters where each cluster has a slice of points and a symbolId of the cluster
func (g *Grid) FindClusters(symbols map[int]*SymbolDef, clusterSize int) []Cluster {
	defMap := make(map[int]*SymbolDef)
	for _, d := range symbols {
		defMap[d.ID] = d
	}

	isWild := func(id int) bool {
		def, exists := defMap[id]
		return exists && len(def.MatchesWith) > 0 && def.MatchesWith[0] == "*"
	}

	isScatter := func(id int) bool {
		def, exists := defMap[id]
		return exists && def.Name == "scatter"
	}

	areCompatible := func(baseID, targetID int) bool {
		if baseID == targetID {
			return true
		}
		if isScatter(baseID) || isScatter(targetID) {
			return false
		}
		baseDef, baseExists := defMap[baseID]
		targetDef, targetExists := defMap[targetID]
		if !baseExists || !targetExists {
			return false
		}
		for _, m := range baseDef.MatchesWith {
			if m == "*" || m == targetDef.Name {
				return true
			}
		}
		for _, m := range targetDef.MatchesWith {
			if m == "*" || m == baseDef.Name {
				return true
			}
		}
		return false
	}

	// Global lock for all symbols (Standard and Wilds)
	visited := make([][]bool, g.Cols)
	for x := 0; x < g.Cols; x++ {
		visited[x] = make([]bool, g.Rows)
	}

	var clusters []Cluster

	// Scan left-to-right, top-to-bottom
	for x := 0; x < g.Cols; x++ {
		for y := 0; y < g.Rows; y++ {
			startID := g.Cells[x][y]

			// Skip if already claimed by a previous cluster, or if it is a Wild
			if visited[x][y] || isWild(startID) {
				continue
			}

			localVisited := make(map[Point]bool)
			queue := []Point{{X: x, Y: y}}
			localVisited[Point{X: x, Y: y}] = true

			var currentCluster []Point
			clusterSymbolID := startID

			for len(queue) > 0 {
				curr := queue[0]
				queue = queue[1:]

				currentCluster = append(currentCluster, Point{X: curr.X, Y: curr.Y})

				dirs := []Point{{X: 0, Y: -1}, {X: 0, Y: 1}, {X: -1, Y: 0}, {X: 1, Y: 0}}
				for _, d := range dirs {
					nx, ny := curr.X+d.X, curr.Y+d.Y

					if nx >= 0 && nx < g.Cols && ny >= 0 && ny < g.Rows {
						np := Point{X: nx, Y: ny}

						// Check local AND global visited state
						if !localVisited[np] && !visited[nx][ny] {
							neighborID := g.Cells[nx][ny]

							if areCompatible(clusterSymbolID, neighborID) {
								localVisited[np] = true
								queue = append(queue, np)
							}
						}
					}
				}
			}

			if len(currentCluster) >= clusterSize {
				// Lock ALL points in the cluster (First Claim)
				for _, p := range currentCluster {
					visited[p.X][p.Y] = true
				}

				clusters = append(clusters, Cluster{
					Points: currentCluster,
					Symbol: *defMap[clusterSymbolID],
				})
			} else {
				// Lock the failed root to prevent redundant BFS cycles
				visited[x][y] = true
			}
		}
	}

	return clusters
}

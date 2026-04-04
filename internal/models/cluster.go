package models

import "sort"

type Cluster struct {
	Points []Point
	Symbol SymbolDef
}

func MergeClusterPoints(clusters []Cluster) []Point {
	var merged []Point

	for _, c := range clusters {
		merged = append(merged, c.Points...)
	}

	sort.Slice(merged, func(i, j int) bool {
		if merged[i].X == merged[j].X {
			return merged[i].Y < merged[j].Y
		}
		return merged[i].X < merged[j].X
	})

	return merged
}

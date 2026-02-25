package features

import (
	"slots/internal/models"
)

type ExpandingWildsFeature struct {
	BaseFeature
	WildID int
}

func NewExpandingWildsFeature(wildId int) *ExpandingWildsFeature {
	return &ExpandingWildsFeature{
		BaseFeature: BaseFeature{
			Type: "EXPANDING_WILDS_FEATURE",
		},
		WildID: wildId,
	}
}

func (f *ExpandingWildsFeature) OnGridEvaluate(ctx FeatureContext) bool {
	g := ctx.GetGrid()
	wildPositions := g.Contain(f.WildID)
	positionsChanged := []models.Point{}
	expandedCols := make(map[int]bool)
	for _, wildPosition := range wildPositions {
		col := wildPosition.X
		if expandedCols[col] {
			continue
		}
		expandedCols[col] = true
		for row := range g.Rows {
			if g.Get(col, row) != f.WildID {
				g.Set(col, row, f.WildID)
				positionsChanged = append(positionsChanged, models.Point{X: col, Y: row})
			}
			// if row != wildPosition.Y {
			// 	g.Set(wildPosition.X, row, f.WildID)
			// 	positionsChanged = append(positionsChanged, grid.Point{X: wildPosition.X, Y: row})
			// }
		}
	}

	if len(positionsChanged) > 0 {
		ctx.PushTimeline(&models.TimelineEvent{
			Type:         "TRANSFORM_FEATURE",
			GridSnapshot: g.Copy(),
			WinAmount:    0,
			Meta: map[string]interface{}{
				"positions": positionsChanged,
				"newId":     f.WildID,
			},
		})
		return true
	}

	return false
}

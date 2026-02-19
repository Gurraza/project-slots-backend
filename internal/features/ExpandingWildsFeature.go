package features

import (
	"slots/internal/grid"
	"slots/internal/timeline"
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
	positionsChanged := []grid.Point{}

	for _, wildPosition := range wildPositions {
		for row := range g.Rows {
			if row != wildPosition.Y {
				g.Set(wildPosition.X, row, f.WildID)
				positionsChanged = append(positionsChanged, grid.Point{X: wildPosition.X, Y: row})
			}
		}
	}

	ctx.PushTimeline(&timeline.TimelineEvent{
		Type:         f.Type,
		GridSnapshot: g.Copy(),
		WinAmount:    0,
		Meta:         positionsChanged,
	})

	return len(positionsChanged) > 0
}

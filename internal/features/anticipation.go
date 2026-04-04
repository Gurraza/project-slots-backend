package features

import "slots/internal/models"

type AnticipationFeature struct {
	BaseFeature
	FeatureSymbolID int
}

func NewAnticipationFeature(prio int, symbolId int) *AnticipationFeature {
	return &AnticipationFeature{
		BaseFeature: BaseFeature{
			Type:     "ANTICIPATION",
			Priority: prio,
		},
		FeatureSymbolID: symbolId,
	}
}

type AnticipationMeta struct {
	Columns            []int           `json:"columns"`
	ID                 int             `json:"id"`
	HighlightPositions []*models.Point `json:"highlightPositions"`
	StartReelIndex     int             `json:"startReelIndex"`
}

func (f *AnticipationFeature) OnGridGenerated(ctx FeatureContext) {
	positions := ctx.GetGrid().Contain(f.FeatureSymbolID)
	for i, pos := range positions {
		if i >= 3 {
			cols := make([]int, ctx.GetGrid().Cols-pos.X)
			for j := range cols {
				cols[j] = pos.X + j
			}
			meta := &AnticipationMeta{
				Columns:            cols,
				ID:                 f.FeatureSymbolID,
				HighlightPositions: positions,
				StartReelIndex:     pos.X,
			}
			ctx.PushTimeline(&models.TimelineEvent{
				Type:         f.Type,
				GridSnapshot: ctx.GetGrid().Copy(),
				Meta:         meta,
			})
			return
		}
	}
}

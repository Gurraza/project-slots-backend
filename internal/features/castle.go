package features

import (
	"slots/internal/models"
)

type CastleFeature struct {
	BaseFeature
	FeatureSymbolID int
	Targets         []int
}

func NewCastleFeature(prio int, symbolId int, targets []int) *CastleFeature {
	return &CastleFeature{
		BaseFeature: BaseFeature{
			Type:     "CASTLE",
			Priority: prio,
		},
		FeatureSymbolID: symbolId,
		Targets:         targets,
	}
}

func (f *CastleFeature) OnGridIdle(ctx FeatureContext) bool {
	positions := ctx.GetGrid().Contain(f.FeatureSymbolID)
	if len(positions) > 0 {
		newId := f.Targets[ctx.GetRandomNumberN(len(f.Targets)-1)]
		for _, pos := range positions {
			ctx.GetGrid().Set(pos.X, pos.Y, newId)
		}
		ctx.PushTimeline(&models.TimelineEvent{
			Type:         "TRANSFORM_FEATURE",
			GridSnapshot: ctx.GetGrid().Copy(),
			WinAmount:    0,
			Meta: map[string]interface{}{
				"positions": positions,
				"newId":     newId,
			},
		})
		return true
	}
	return false
}

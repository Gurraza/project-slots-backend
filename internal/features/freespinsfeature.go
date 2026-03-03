package features

import (
	"slots/internal/models"
)

type FreeSpinsFeature struct {
	BaseFeature
	FeatureSymbolID int
	FreeSpins       int
	Features        []GameFeature
}

func NewFreeSpinsFeature(sid int, freespins int, features []GameFeature) *FreeSpinsFeature {
	return &FreeSpinsFeature{
		BaseFeature: BaseFeature{
			Type: "FREE_SPINS_FEATURE",
		},
		FeatureSymbolID: sid,
		Features:        features,
		FreeSpins:       freespins,
	}
}

// func (f *FreeSpinsFeature) GetSymbols(ctx FeatureContext) []*models.SymbolDef {
// 	return []*models.SymbolDef{f.FeatureSymbol}
// }

func (f *FreeSpinsFeature) OnGridIdle(ctx FeatureContext) bool {
	if len(ctx.GetGrid().Contain(f.FeatureSymbolID)) >= 3 {
		ctx.PushTimeline(&models.TimelineEvent{
			Type: "BONUS_GAME",
		})
		for _, newFeature := range f.Features {
			ctx.AddFeature(newFeature)
			newFeature.Init(ctx)
		}
		ctx.RemoveFeature(f.Type)
		for freeSpinsLeft := range f.FreeSpins {
			ctx.PushTimeline(&models.TimelineEvent{
				Type:         f.Type,
				GridSnapshot: ctx.GetGrid().Copy(),
				WinAmount:    0,
				Meta:         freeSpinsLeft,
			})
			ctx.Spin()
		}
		ctx.AddFeature(f)

		for _, newFeature := range f.Features {
			ctx.RemoveFeature(newFeature.GetType())
		}
		return false
	}
	return false
}

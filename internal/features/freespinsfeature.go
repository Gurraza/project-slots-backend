package features

import (
	"slots/internal/symbol"
	"slots/internal/timeline"
)

type FreeSpinsFeature struct {
	BaseFeature
	FeatureSymbol *symbol.SymbolDef
	FreeSpins     int
	Features      []GameFeature
}

func NewFreeSpinsFeature(s *symbol.SymbolDef, freespins int, features []GameFeature) *FreeSpinsFeature {
	return &FreeSpinsFeature{
		BaseFeature: BaseFeature{
			Type: "FREE_SPINS_FEATURE",
		},
		FeatureSymbol: s,
		Features:      features,
		FreeSpins:     freespins,
	}
}

func (f *FreeSpinsFeature) GetSymbols(ctx FeatureContext) []*symbol.SymbolDef {
	return []*symbol.SymbolDef{f.FeatureSymbol}
}

func (f *FreeSpinsFeature) OnGridIdle(ctx FeatureContext) bool {
	if len(ctx.GetGrid().Contain(f.FeatureSymbol.ID)) >= 3 {
		// ctx.PushTimeline(&timeline.TimelineEvent{
		// 	Type:         "FREE_SPINS_FEATURE",
		// 	GridSnapshot: ctx.GetGrid().Copy(),
		// })
		for _, newFeature := range f.Features {
			ctx.AddFeature(newFeature)
			newFeature.Init(ctx)
		}
		ctx.RemoveFeature(f.Type)
		for freeSpinsLeft := range f.FreeSpins {
			ctx.PushTimeline(&timeline.TimelineEvent{
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

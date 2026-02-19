package features

import "slots/internal/symbol"

type WildFeature struct {
	BaseFeature
	FeatureSymbol *symbol.SymbolDef
}

func NewWildFeature(s *symbol.SymbolDef) *WildFeature {
	return &WildFeature{
		BaseFeature: BaseFeature{
			Type: "WILD_FEATURE",
		},
		FeatureSymbol: s,
	}
}

func (f *WildFeature) Init(ctx FeatureContext) {
	ctx.AddSymbol(f.FeatureSymbol)
}

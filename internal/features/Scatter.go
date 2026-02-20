package features

import "slots/internal/symbol"

type ScatterFeature struct {
	BaseFeature
	FeatureSymbol *symbol.SymbolDef
}

func NewScatterFeature(s *symbol.SymbolDef) *ScatterFeature {
	return &ScatterFeature{
		BaseFeature: BaseFeature{
			Type: "SCATTER_FEATURE",
		},
		FeatureSymbol: s,
	}
}

func (f *ScatterFeature) GetSymbols(ctx FeatureContext) []*symbol.SymbolDef {
	return []*symbol.SymbolDef{f.FeatureSymbol}
}

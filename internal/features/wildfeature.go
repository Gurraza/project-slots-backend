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

func (f *WildFeature) GetSymbols(ctx FeatureContext) []*symbol.SymbolDef {
	return []*symbol.SymbolDef{f.FeatureSymbol}
}

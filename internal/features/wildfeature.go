package features

import (
	"slots/internal/models"
)

type WildFeature struct {
	BaseFeature
	FeatureSymbol *models.SymbolDef
}

func NewWildFeature(s *models.SymbolDef) *WildFeature {
	return &WildFeature{
		BaseFeature: BaseFeature{
			Type: "WILD_FEATURE",
		},
		FeatureSymbol: s,
	}
}

func (f *WildFeature) GetSymbols(ctx FeatureContext) []*models.SymbolDef {
	return []*models.SymbolDef{f.FeatureSymbol}
}

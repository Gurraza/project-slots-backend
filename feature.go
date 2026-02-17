package main

// Feature is the interface all game mechanics must implement
type GameFeature interface {
	Init(config *GameConfig)
	GetSymbols() []SymbolDef

	OnSpinStart(*RoundSession)
	OnGridEvaluate(*RoundSession) bool
	OnGridIdle(*RoundSession) bool
	OnSpinEnd(*RoundSession)
}

type BaseFeature struct {
	FeatureName   string
	FeatureSymbol *SymbolDef
	RoundSession  *RoundSession
}

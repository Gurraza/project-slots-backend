package main

// Feature is the interface all game mechanics must implement
type GameFeature interface {
	Init()
	GetSymbols() []SymbolDef

	OnSpinStart(*GameState)
	OnGridEvaluate(*GameState) bool
	OnGridIdle(*GameState) bool
	OnSpinEnd(*GameState)
}

type BaseFeature struct {
	Type string
}

// --- Implementing the Interface Methods ---

// Init: Sets up the initial configuration
func (f *BaseFeature) Init() {
}

// GetSymbols: Returns the symbols this feature uses
func (f *BaseFeature) GetSymbols() []SymbolDef {
	// Assuming your Config has a list of all symbols
	return []SymbolDef{}
}

// OnSpinStart: Reset state before the reels move
func (f *BaseFeature) OnSpinStart(session *GameState) {
	// Logic to run right before spin (e.g., deducting balance)
}

// OnGridEvaluate: Logic to change the grid (e.g., Wild expansion)
func (f *BaseFeature) OnGridEvaluate(session *GameState) bool {
	// Return true if you modified the grid and need a re-evaluation
	return false
}

func (f *BaseFeature) OnGridIdle(session *GameState) bool {
	return false
}

func (f *BaseFeature) OnSpinEnd(session *GameState) {
	// Post-game cleanup or stats logging
}

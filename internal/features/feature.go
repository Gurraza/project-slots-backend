package features

import (
	"slots/internal/grid"
	"slots/internal/symbol"
	"slots/internal/timeline"
)

type FeatureContext interface {
	GetGrid() *grid.Grid
	SetGrid(*grid.Grid)
	GetSymbols() []*symbol.SymbolDef
	AddSymbol(symbolDef *symbol.SymbolDef)
	GetRandomSymbol(grid *grid.Grid, col int, row int) *symbol.SymbolDef
	PushTimeline(timelineEvent *timeline.TimelineEvent)
}

// Feature is the interface all game mechanics must implement
type GameFeature interface {
	Init(FeatureContext)

	OnSpinStart(FeatureContext)
	OnGridEvaluate(FeatureContext) bool
	OnGridIdle(FeatureContext) bool
	OnSpinEnd(FeatureContext)
}

type BaseFeature struct {
	Type string
}

// --- Implementing the Interface Methods ---

// Init: Sets up the initial configuration
func (f *BaseFeature) Init(ctx FeatureContext) {
}

// OnSpinStart: Reset state before the reels move
func (f *BaseFeature) OnSpinStart(ctx FeatureContext) {
	// Logic to run right before spin (e.g., deducting balance)
}

// OnGridEvaluate: Logic to change the grid (e.g., Wild expansion)
func (f *BaseFeature) OnGridEvaluate(ctx FeatureContext) bool {
	// Return true if you modified the grid and need a re-evaluation
	return false
}

func (f *BaseFeature) OnGridIdle(ctx FeatureContext) bool {
	return false
}

func (f *BaseFeature) OnSpinEnd(ctx FeatureContext) {
	// Post-game cleanup or stats logging
}

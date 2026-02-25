package features

import (
	"slots/internal/models"
)

type FeatureContext interface {
	GetGrid() *models.Grid
	SetGrid(*models.Grid)
	GetSymbols() map[int]*models.SymbolDef
	AddSymbol(symbolDef *models.SymbolDef)
	GetRandomSymbol(grid *models.Grid, col int, row int) *models.SymbolDef
	PushTimeline(timelineEvent *models.TimelineEvent)
	AddFeature(newFeature GameFeature)
	RemoveFeature(featureType string)
	Spin() []*models.TimelineEvent
}

// Feature is the interface all game mechanics must implement
type GameFeature interface {
	Init(FeatureContext)
	GetType() string

	OnSpinStart(FeatureContext)
	OnGridEvaluate(FeatureContext) bool
	OnGridIdle(FeatureContext) bool
	OnSpinEnd(FeatureContext)
	GetSymbols(FeatureContext) []*models.SymbolDef
}

type BaseFeature struct {
	Type string
}

// --- Implementing the Interface Methods ---

func (f *BaseFeature) GetSymbols(ctx FeatureContext) []*models.SymbolDef {
	return []*models.SymbolDef{}
}

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

func (f *BaseFeature) GetType() string {
	return f.Type
}

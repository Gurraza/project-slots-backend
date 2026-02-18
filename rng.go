package main

import (
	"math/rand"
)

// RNG interface
type RNG interface {
	Uint32() uint32
	Range(max int) int
	GetRandomSymbol(gameState *GameState) SymbolDef
}

// GoRNG wraps math/rand for reproducibility
type GoRNG struct {
	r *rand.Rand
}

// Constructor
func NewGoRNG(seed int64) *GoRNG {
	return &GoRNG{r: rand.New(rand.NewSource(seed))}
}

func (g *GoRNG) Uint32() uint32 {
	return g.r.Uint32()
}

func (g *GoRNG) Range(max int) int {
	return g.r.Intn(max)
}

func (g *GoRNG) GetRandomSymbol(gameState *GameState) SymbolDef {

	totalWeight := 0
	for i := range gameState.Config.Symbols {
		totalWeight += gameState.Config.Symbols[i].GetWeight(gameState)
	}
	randomNmr := gameState.RNG.Range(totalWeight)

	for _, s := range gameState.Config.Symbols {
		w := s.GetWeight(gameState)
		if randomNmr < w {
			return s
		}
		randomNmr -= w
	}

	// will never happen
	return gameState.Config.Symbols[0]
}

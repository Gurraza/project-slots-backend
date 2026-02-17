package main

import (
	"math/rand"
)

// RNG interface
type RNG interface {
	Uint32() uint32
	Range(max int) int
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

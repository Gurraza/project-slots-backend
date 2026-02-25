package models

import (
	"fmt"
	"math/rand"
)

// RNG interface
type RNG interface {
	Uint32() uint32
	Range(max int) int
	Intn(n int) int
	GetSeed() int64
}

// GoRNG wraps math/rand for reproducibility
type GoRNG struct {
	r    *rand.Rand
	Seed int64 `json:"seed"`
}

// Constructor
func NewGoRNG(seed int64) *GoRNG {
	fmt.Println(seed)
	return &GoRNG{
		r:    rand.New(rand.NewSource(seed)),
		Seed: seed,
	}
}

func (g *GoRNG) Uint32() uint32 {
	return g.r.Uint32()
}

func (g *GoRNG) Intn(n int) int {
	return g.r.Intn(n)
}

func (g *GoRNG) Range(max int) int {
	return g.r.Intn(max)
}

func (g *GoRNG) GetSeed() int64 {
	return g.Seed
}

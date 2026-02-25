package models

type SymbolDef struct {
	ID           int
	Name         string
	WeightConfig WeightConfig
	MatchesWith  []string
	Payouts      []float64
}

type WeightModifier interface {
	Apply(currentWeight int, ctx *WeightContext, s *SymbolDef) int
}

type WeightConfig struct {
	FixedWeight int
	Modifiers   []WeightModifier
}

type WeightContext struct {
	ReelIndex int
	RowIndex  int
	Grid      *Grid
}

func (s *SymbolDef) GetWeight(ctx *WeightContext) int {
	w := s.WeightConfig.FixedWeight
	for _, mod := range s.WeightConfig.Modifiers {
		w = mod.Apply(w, ctx, s)
		if w <= 0 {
			return 0
		}
	}
	return w
}

type CountWeight struct {
	// Scales is a list of percentage factors (e.g., 100 = 1.0x, 50 = 0.5x)
	// Index 0 = 0 previous instances, Index 1 = 1 previous instance, etc.
	Scales []int
}

func (m *CountWeight) Apply(currentWeight int, ctx *WeightContext, symbolDef *SymbolDef) int {
	// 1. Get how many times this symbol has already spawned
	// (You would pass the SymbolID into the modifier or Context,
	// strictly speaking the modifier might need to know which symbol it belongs to,
	// or we assume the modifier is attached to the specific symbol instance).
	// For this generic example, let's assume we look up the count externally or pass ID.
	// Let's assume the Context tracks the ID we are currently calculating for.

	// NOTE: In a real engine, you'd pass the target SymbolID to Apply as well.

	count := len(ctx.Grid.Contain(symbolDef.ID))

	if count >= len(m.Scales) {
		// Fallback: If we exceed the array, use the last value or 0?
		// Professional slot engines usually default to the last defined value.
		return (currentWeight * m.Scales[len(m.Scales)-1]) / 100
	}

	scaleFactor := m.Scales[count]

	// Integer math: (Weight * Factor) / 100
	// e.g. (1000 * 50) / 100 = 500
	return (currentWeight * scaleFactor) / 100
}

type ReelWeight struct {
	// Multipliers for each reel index.
	// e.g. [0, 0, 100, 200] for 0x, 0x, 1x, 2x
	ReelMultipliers []int
}

func (m *ReelWeight) Apply(currentWeight int, ctx *WeightContext, symbolDef *SymbolDef) int {
	if ctx.ReelIndex >= len(m.ReelMultipliers) {
		return currentWeight // Safe fallback
	}

	factor := m.ReelMultipliers[ctx.ReelIndex]

	// Standard scaling: (Weight * Factor) / 100
	// If factor is 200 (2x), and base is 50 -> returns 100.
	return (currentWeight * factor) / 100
}

type SameReelWeight struct {
	TargetSymbolID int
	Factor         int // e.g. 50 for 50%
}

func (m *SameReelWeight) Apply(currentWeight int, ctx *WeightContext, symbolDef *SymbolDef) int {
	// Check if the target symbol is already in the current reel's generated column
	// (Assuming GridState is populated row by row or col by col)
	present := false
	for row := 0; row < ctx.Grid.Rows; row++ {
		if ctx.Grid.Get(ctx.ReelIndex, row) == m.TargetSymbolID {
			present = true
			break
		}
	}

	if present {
		return (currentWeight * m.Factor) / 100
	}
	return currentWeight
}

func (s1 *SymbolDef) Compatible(s2 *SymbolDef) bool {
	if s1 == nil || s2 == nil {
		return false
	}

	// Same symbol always matches
	if s1.ID == s2.ID {
		return true
	}

	// Check if s1 allows s2
	for _, name := range s1.MatchesWith {
		if name == s2.Name || name == "*" {
			return true
		}
	}

	// Check if s2 allows s1 (the missing half today)
	for _, name := range s2.MatchesWith {
		if name == s1.Name || name == "*" {
			return true
		}
	}

	return false
}

func (s *SymbolDef) IsWild() bool {
	for _, m := range s.MatchesWith {
		if m == "*" {
			return true
		}
	}
	return false
}

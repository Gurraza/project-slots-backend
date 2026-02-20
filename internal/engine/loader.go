package engine

import (
	"encoding/json"
	"errors"
	"slots/internal/features"
	"slots/internal/symbol"
)

// --- 1. Define Flat DTOs for JSON Parsing ---

type jsonConfigDTO struct {
	Cols     int              `json:"Cols"`
	Rows     int              `json:"Rows"`
	Symbols  []jsonSymbolDTO  `json:"Symbols"`
	Features []jsonFeatureDTO `json:"Features"`
}

type jsonSymbolDTO struct {
	ID           int                 `json:"ID"`
	Name         string              `json:"Name"`
	Payouts      []float64           `json:"Payouts"`
	MatchesWith  []string            `json:"MatchesWith"`
	WeightConfig jsonWeightConfigDTO `json:"WeightConfig"`
}

type jsonWeightConfigDTO struct {
	FixedWeight int                `json:"FixedWeight"`
	Modifiers   []jsonWeightModDTO `json:"Modifiers"`
}

type jsonWeightModDTO struct {
	Type            string `json:"Type"`
	Scales          []int  `json:"Scales,omitempty"`
	ReelMultipliers []int  `json:"ReelMultipliers,omitempty"`
	TargetSymbolID  int    `json:"TargetSymbolID,omitempty"`
	Factor          int    `json:"Factor,omitempty"`
}

type jsonFeatureDTO struct {
	Type             string        `json:"Type"`
	TargetSymbolID   int           `json:"TargetSymbolID,omitempty"`
	FeatureSymbol    jsonSymbolDTO `json:"FeatureSymbol"`
	Paylines         [][]int       `json:"Paylines,omitempty"`
	ExcludeSymbolIDs []int         `json:"ExcludeSymbolIDs,omitempty"`
}

func LoadConfigFromJSON(jsonData []byte) (*GameConfig, error) {
	var dto jsonConfigDTO
	if err := json.Unmarshal(jsonData, &dto); err != nil {
		return nil, err
	}

	config := &GameConfig{
		Cols:     dto.Cols,
		Rows:     dto.Rows,
		Symbols:  make([]*symbol.SymbolDef, 0, len(dto.Symbols)),
		Features: make([]features.GameFeature, 0, len(dto.Features)),
	}

	// Helper map to quickly find symbols by ID for feature injection
	symbolLookup := make(map[int]*symbol.SymbolDef)

	// A. Map Symbols and Modifiers
	for _, sDTO := range dto.Symbols {
		sym := SymbolFronJson(sDTO)
		config.Symbols = append(config.Symbols, sym)
		symbolLookup[sym.ID] = sym
	}

	// B. Map Features via Factory
	for _, fDTO := range dto.Features {
		switch fDTO.Type {
		case "CLUSTER_FEATURE":
			config.Features = append(config.Features, features.NewClusterFeature())
		case "WILD_FEATURE":
			targetSym, exists := symbolLookup[fDTO.TargetSymbolID]
			if !exists {
				return nil, errors.New("WILD_FEATURE references missing symbol ID")
			}
			config.Features = append(config.Features, features.NewWildFeature(targetSym))
		case "PAYLINES_FEATURE":

			config.Features = append(config.Features, features.NewPaylineFeature(fDTO.Paylines, fDTO.ExcludeSymbolIDs))
		case "EXPANDING_WILDS_FEATURE":
			config.Features = append(config.Features, features.NewExpandingWildsFeature(fDTO.TargetSymbolID))
		case "SCATTER_FEATURE":
			s := SymbolFronJson(fDTO.FeatureSymbol)
			config.Features = append(config.Features, features.NewScatterFeature(s))
		}
	}

	return config, nil
}

func SymbolFronJson(sDTO jsonSymbolDTO) *symbol.SymbolDef {
	sym := &symbol.SymbolDef{
		ID:          sDTO.ID,
		Name:        sDTO.Name,
		Payouts:     sDTO.Payouts,
		MatchesWith: sDTO.MatchesWith,
		WeightConfig: symbol.WeightConfig{
			FixedWeight: sDTO.WeightConfig.FixedWeight,
			Modifiers:   make([]symbol.WeightModifier, 0),
		},
	}

	// Map Modifiers via Factory
	for _, mDTO := range sDTO.WeightConfig.Modifiers {
		switch mDTO.Type {
		case "CountWeight":
			sym.WeightConfig.Modifiers = append(sym.WeightConfig.Modifiers, &symbol.CountWeight{Scales: mDTO.Scales})
		case "ReelWeight":
			sym.WeightConfig.Modifiers = append(sym.WeightConfig.Modifiers, &symbol.ReelWeight{ReelMultipliers: mDTO.ReelMultipliers})
		case "SameReelWeight":
			sym.WeightConfig.Modifiers = append(sym.WeightConfig.Modifiers, &symbol.SameReelWeight{TargetSymbolID: mDTO.TargetSymbolID, Factor: mDTO.Factor})
		}
	}

	return sym
}

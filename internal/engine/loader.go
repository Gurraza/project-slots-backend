package engine

import (
	"encoding/json"
	"slots/internal/features"
	"slots/internal/models"
	"strconv"
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
	Type             string           `json:"Type"`
	Priority         int              `json:"Priority"`
	TargetSymbolID   int              `json:"TargetSymbolID,omitempty"`
	FeatureSymbol    jsonSymbolDTO    `json:"FeatureSymbol"`
	Paylines         [][]int          `json:"Paylines,omitempty"`
	ExcludeSymbolIDs []int            `json:"ExcludeSymbolIDs,omitempty"`
	FreeSpins        int              `json:"FreeSpins,omitempty"`
	Features         []jsonFeatureDTO `json:"Features,omitempty"`
	ClusterSize      int              `json:"clusterSize,omitempty"`
	Targets          []int            `json:"targets,omitempty"`
	TargetMap        map[string]int   `json:"TargetMap,omitempty"`
}

func LoadConfigFromJSON(jsonData []byte) (*GameConfig, error) {
	var dto jsonConfigDTO
	if err := json.Unmarshal(jsonData, &dto); err != nil {
		return nil, err
	}

	config := &GameConfig{
		Cols:     dto.Cols,
		Rows:     dto.Rows,
		Symbols:  make([]*models.SymbolDef, 0, len(dto.Symbols)),
		Features: make([]features.GameFeature, 0, len(dto.Features)),
	}

	// Helper map to quickly find symbols by ID for feature injection
	symbolLookup := make(map[int]*models.SymbolDef)

	// A. Map Symbols and Modifiers
	for _, sDTO := range dto.Symbols {
		sym := SymbolFromJSON(sDTO)
		config.Symbols = append(config.Symbols, sym)
		symbolLookup[sym.ID] = sym
	}

	// B. Map Features via Factory
	for _, fDTO := range dto.Features {
		f := FeatureFromJSON(fDTO, symbolLookup)
		if f != nil {
			config.Features = append(config.Features, f)
		}
	}

	return config, nil
}

func FeatureFromJSON(fDTO jsonFeatureDTO, symbolLookup map[int]*models.SymbolDef) features.GameFeature {
	switch fDTO.Type {
	case "CLUSTER_FEATURE":
		return features.NewClusterFeature(fDTO.Priority, fDTO.ClusterSize, fDTO.ExcludeSymbolIDs)
	case "WILD_FEATURE":
		return features.NewWildFeature(fDTO.Priority, SymbolFromJSON(fDTO.FeatureSymbol))
	case "PAYLINES_FEATURE":
		return features.NewPaylineFeature(fDTO.Priority, fDTO.Paylines, fDTO.ExcludeSymbolIDs)
	case "CASTLE":
		return features.NewCastleFeature(fDTO.Priority, fDTO.TargetSymbolID, fDTO.Targets)
	case "EXPANDING_WILDS_FEATURE":
		return features.NewExpandingWildsFeature(fDTO.Priority, fDTO.TargetSymbolID)
	case "ANTICIPATION":
		return features.NewAnticipationFeature(fDTO.Priority, fDTO.TargetSymbolID)
	case "ANYWHERE_CASCADE_FEATURE":
		return features.NewAnywhereCascadeFeature(fDTO.Priority, fDTO.ClusterSize, fDTO.Targets)
	case "FREE_SPINS_FEATURE":
		fs := make([]features.GameFeature, 0, len(fDTO.Features))
		for _, f := range fDTO.Features {
			nestedFeature := FeatureFromJSON(f, symbolLookup)
			if nestedFeature != nil {
				fs = append(fs, nestedFeature)
			}
		}
		return features.NewFreeSpinsFeature(fDTO.Priority, fDTO.TargetSymbolID, fDTO.FreeSpins, fs)
	case "TRIBUTE_HARVEST":
		// Convert map[string]int to map[int]int
		parsedTargets := make(map[int]int)
		for k, v := range fDTO.TargetMap {
			intKey, err := strconv.Atoi(k)
			if err != nil {
				panic("Invalid TargetMap key in TRIBUTE_HARVEST JSON, must be integer string")
			}
			parsedTargets[intKey] = v
		}
		return features.NewTributeHarvestFeature(fDTO.Priority, fDTO.ClusterSize, parsedTargets)
	}

	panic("Feature type unrecognized in loader.go, FeatureFromJSON")
}

func SymbolFromJSON(sDTO jsonSymbolDTO) *models.SymbolDef {
	sym := &models.SymbolDef{
		ID:          sDTO.ID,
		Name:        sDTO.Name,
		Payouts:     sDTO.Payouts,
		MatchesWith: sDTO.MatchesWith,
		WeightConfig: models.WeightConfig{
			FixedWeight: sDTO.WeightConfig.FixedWeight,
			Modifiers:   make([]models.WeightModifier, 0),
		},
	}

	// Map Modifiers via Factory
	for _, mDTO := range sDTO.WeightConfig.Modifiers {
		switch mDTO.Type {
		case "CountWeight":
			sym.WeightConfig.Modifiers = append(sym.WeightConfig.Modifiers, &models.CountWeight{Scales: mDTO.Scales})
		case "ReelWeight":
			sym.WeightConfig.Modifiers = append(sym.WeightConfig.Modifiers, &models.ReelWeight{ReelMultipliers: mDTO.ReelMultipliers})
		case "SameReelWeight":
			sym.WeightConfig.Modifiers = append(sym.WeightConfig.Modifiers, &models.SameReelWeight{TargetSymbolID: mDTO.TargetSymbolID, Factor: mDTO.Factor})
		}
	}

	return sym
}

package features

import (
	"slots/internal/models"
)

type TributeHarvestFeature struct {
	BaseFeature
	Targets     map[int]int
	ClusterSize int
}

func NewTributeHarvestFeature(prio int, clusterSize int, targets map[int]int) *TributeHarvestFeature {
	return &TributeHarvestFeature{
		BaseFeature: BaseFeature{
			Type:     "TRIBUTE_HARVEST",
			Priority: prio,
		},
		ClusterSize: clusterSize,
		Targets:     targets,
	}
}

func (f *TributeHarvestFeature) OnGridIdle(ctx FeatureContext) bool {
	clustersFound := ctx.GetGrid().FindClusters(ctx.GetSymbols(), f.ClusterSize)
	if len(clustersFound) == 0 {
		return false
	}
	resourcesToSuck := make([]models.Point, 0)
	activeResourceIDs := make(map[int]bool)
	clusters := make([]models.Cluster, 0)
	var superIdPos models.Point
	for _, cluster := range clustersFound {
		for _, p := range cluster.Points {
			symbolID := ctx.GetGrid().Get(p.X, p.Y)
			if resourceID, exists := f.Targets[symbolID]; exists {
				clusters = append(clusters, cluster)
				superIdPos = p
				activeResourceIDs[resourceID] = true
				break
			}
		}
	}

	if len(clusters) == 0 {
		return false
	}

	grid := ctx.GetGrid()
	for x := range grid.Cols {
		for y := range grid.Rows {
			currentID := grid.Get(x, y)
			// 2. Map symbolID to resourceID before checking active status
			// if resourceID, exists := f.Targets[currentID]; exists && activeResourceIDs[resourceID] {
			// 	resourcesToSuck = append(resourcesToSuck, models.Point{X: x, Y: y})
			// }
			if activeResourceIDs[currentID] {
				resourcesToSuck = append(resourcesToSuck, models.Point{X: x, Y: y})
			}
		}
	}

	merged := models.MergeClusterPoints(clusters)
	merged = append(merged, resourcesToSuck...)

	explosions := ctx.GetGrid().ExplodeAndCascade(merged, func(x, y int) *models.SymbolDef {
		return ctx.GetRandomSymbol(ctx.GetGrid(), x, y)
	})

	ctx.PushTimeline(&models.TimelineEvent{
		Type:         "TRIBUTE_HARVEST",
		GridSnapshot: ctx.GetGrid().Copy(),
		Meta: map[string]interface{}{
			// Send the unified structure to the frontend
			"source":          superIdPos,
			"resourcesToSuck": resourcesToSuck,
		},
	})

	spinWinAmount := CalculatePayout(clusters)

	ctx.PushTimeline(&models.TimelineEvent{
		Type:         "EXPLODE_AND_CASCADE_FEATURE",
		GridSnapshot: ctx.GetGrid().Copy(),
		WinAmount:    spinWinAmount,
		Meta: map[string]interface{}{
			// Send the unified structure to the frontend
			"explosions":      explosions,
			"resourcesToSuck": resourcesToSuck,
		},
	})
	return true
}

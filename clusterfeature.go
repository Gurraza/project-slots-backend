package main

type ClusterFeature struct {
	BaseFeature
}

func NewClusterFeature() *ClusterFeature {
	return &ClusterFeature{
		BaseFeature: BaseFeature{
			Type: "CLUSTER_FEATURE",
		},
	}
}

func (f *ClusterFeature) OnGridIdle(gameState *GameState) bool {
	cluster := FindClusters(*gameState.Grid, gameState.Config.Symbols)
	mergedCluster := make([]Point, 0)
	for _, c := range cluster {
		mergedCluster = append(mergedCluster, c...)
	}
	if len(mergedCluster) == 0 {
		return false
	}

	replacements := make([]int, len(mergedCluster))
	for i := range replacements {
		replacements[i] = gameState.RNG.GetRandomSymbol(gameState).ID
	}
	gameState.Grid, _ = ExplodeAndCascade(*gameState.Grid, mergedCluster, replacements)

	gameState.Timeline = append(gameState.Timeline, TimelineEvent{
		Type:         "ExplodeAndCascade",
		GridSnapshot: gameState.Grid.Copy(),
		WinAmount:    0,
		Meta: map[string]interface{}{
			"points":       mergedCluster,
			"replacements": replacements,
		},
	})
	return true
}

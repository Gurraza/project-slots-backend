package main

type clusterFeature struct {
	BaseFeature
}

func (f *clusterFeature) OnGridIdle(gameState *GameState) bool {

	return false
}

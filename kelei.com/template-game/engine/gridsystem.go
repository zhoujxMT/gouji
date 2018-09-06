package engine

type GridSystem struct {
	grids []*Grid
}

func NewGridSystem() *GridSystem {
	gridSystem := GridSystem{}
	gridSystem.grids = []*Grid{}
	return &gridSystem
}

func (this *GridSystem) AddGrid(gridType int, theirUser *User) {
	grid := NewGrid(gridType, theirUser)
	this.grids = append(this.grids, grid)
}

package engine

func (this *Turret) getIndex() int {
	return this.index
}

func (this *Turret) setIndex(index int) {
	this.index = index
}

func (this *Turret) getLevel() int {
	return this.level
}

func (this *Turret) setLevel(level int) {
	this.level = level
}

func (this *Turret) getGrid() *Grid {
	return this.grid
}

func (this *Turret) setGrid(grid *Grid) {
	this.grid = grid
}

func (this *Turret) getImage() *TurretImage {
	image := TurretImage{}
	image.I = this.getIndex()
	image.Level = this.getLevel()
	return &image
}

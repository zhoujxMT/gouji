package engine

import (
	"encoding/json"

	"kelei.com/utils/logger"
)

type TurretSystem struct {
	turrentCount int
	turrets      []*Turret
	room         *Room
}

type TurretSystemImage struct {
	Turrets []TurretImage
}

func NewTurretSystem(room *Room) *TurretSystem {
	turretSystem := TurretSystem{}
	turretSystem.turrentCount = 4
	turretSystem.turrets = []*Turret{}
	turretSystem.room = room
	turretSystem.build()
	return &turretSystem
}

func (this *TurretSystem) build() {
	for i := 0; i < this.turrentCount; i++ {
		this.AddTurret(i + 1)
	}
}

func (this *TurretSystem) AddTurret(level int) {
	turret := NewTurret()
	turret.setIndex(len(this.turrets))
	turret.setLevel(level)
	this.turrets = append(this.turrets, turret)
}

func (this *TurretSystem) getTurrets() []*Turret {
	return this.turrets
}

func (this *TurretSystem) getRoom() *Room {
	return this.room
}

func (this *TurretSystem) fill() {
	this.bindToGrid()
}

func (this *TurretSystem) bindToGrid() {
	gridSystem := this.getRoom().GetGridSystem()
	grids := gridSystem.getGrids()
	index := 0
	for _, grid := range grids {
		if grid.isTurret() {
			turret := this.getTurrets()[index]
			grid.setTurret(turret)
			turret.setGrid(grid)
			index++
		}
	}
}

func (this *TurretSystem) getImage() *string {
	image := TurretSystemImage{[]TurretImage{}}
	turrets := this.getTurrets()
	for _, turret := range turrets {
		turretImage := *turret.getImage()
		image.Turrets = append(image.Turrets, turretImage)
	}
	b, err := json.Marshal(image)
	logger.CheckFatal(err)
	msg := string(b)
	return &msg
}

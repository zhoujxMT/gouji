package engine

type Turret struct {
	index int
	level int
	grid  *Grid
}

type TurretImage struct {
	I     int
	Level int
}

func NewTurret() *Turret {
	turret := &Turret{}
	return turret
}

func (this *Turret) attack() {

}

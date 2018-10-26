package engine

/*
	推送开赛的所有信息
*/
func (this *Room) pushAll() {
	this.pushGrid()
	this.pushUser()
	this.pushTurret()
}

/*
	推送网格信息
*/
func (this *Room) pushGrid() {
	gridSystem := this.GetGridSystem()
	image := gridSystem.getImage()
	this.push("GridImage", image)
}

/*
	推送玩家信息
*/
func (this *Room) pushUser() {
	roomUserSystem := this.GetRoomUserSystem()
	image := roomUserSystem.getImage()
	this.push("UserImage", image)
}

/*
	推送炮塔信息
*/
func (this *Room) pushTurret() {
	turretSystem := this.GetTurretSystem()
	image := turretSystem.getImage()
	this.push("TurretImage", image)
}

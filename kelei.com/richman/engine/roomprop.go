package engine

func (this *Room) GetRoomID() string {
	return this.roomid
}

func (this *Room) setRoomID(roomid string) {
	this.roomid = roomid
}

func (this *Room) GetStatus() int {
	return this.status
}

func (this *Room) setStatus(status int) {
	this.status = status
}

func (this *Room) GetUsers() []*User {
	return this.users
}

func (this *Room) GetGridSystem() *GridSystem {
	return this.gridSystem
}

func (this *Room) GetRoomUserSystem() *RoomUserSystem {
	return this.roomUserSystem
}

func (this *Room) GetItemSystem() *ItemSystem {
	return this.itemSystem
}

func (this *Room) GetTurretSystem() *TurretSystem {
	return this.turretSystem
}

func (this *Room) GetPCount() int {
	return this.pcount
}

func (this *Room) ISMatching() bool {
	return this.GetStatus() == STATUS_MATCHING
}

func (this *Room) ISSetout() bool {
	return this.GetStatus() == STATUS_SETOUT
}

package engine

import (
	"encoding/json"
	"fmt"
)

func (this *Grid) getIndex() int {
	return this.index
}

func (this *Grid) setIndex(index int) {
	this.index = index
}

func (this *Grid) getGridType() int {
	return this.gridType
}

func (this *Grid) SetGridType(gridType int) {
	this.gridType = gridType
}

func (this *Grid) getTurret() *Turret {
	return this.turret
}

func (this *Grid) setTurret(turret *Turret) {
	this.turret = turret
}

func (this *Grid) isTurret() bool {
	return this.getGridType() == GRID_TYPE_TURRET
}

func (this *Grid) getTheirUser() *User {
	return this.theirUser
}

func (this *Grid) SetTheirUser(theirUser *User) {
	this.theirUser = theirUser
}

func (this *Grid) getItem() *Item {
	return this.item
}

func (this *Grid) setItem(item *Item) {
	this.item = item
}

func (this *Grid) getUser() *User {
	return this.user
}

func (this *Grid) setUser(user *User) {
	this.user = user
}

func (this *Grid) getRoom() *Room {
	return this.room
}

func (this *Grid) setRoom(room *Room) {
	this.room = room
}

func (this *Grid) showImage() {
	image, _ := json.Marshal(*this.getImage())
	fmt.Println(string(image))
}

func (this *Grid) getImage() *GridImage {
	image := GridImage{}
	image.I = this.getIndex()
	image.GridType = this.getGridType()
	if this.getTheirUser() != nil {
		image.TheirUser = this.getTheirUser().getUserID2Int()
	}
	if this.getItem() != nil {
		image.Item = this.getItem().getItemID()
	}
	if this.getUser() != nil {
		image.User = this.getUser().getUserID2Int()
	}
	return &image
}

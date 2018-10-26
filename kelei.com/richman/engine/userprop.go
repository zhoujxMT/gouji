package engine

import (
	"net"
	"strconv"

	"kelei.com/utils/common"
)

func (this *User) getUserID() string {
	return this.userid
}

func (this *User) getUserID2Int() int {
	userid, _ := strconv.Atoi(this.getUserID())
	return userid
}

func (this *User) setUserID(userid string) {
	this.userid = userid
}

func (this *User) getConn() net.Conn {
	return this.conn
}

func (this *User) setConn(conn net.Conn) {
	this.conn = conn
}

func (this *User) getIndex() int {
	return this.index
}

func (this *User) setIndex(index int) {
	this.index = index
}

func (this *User) getGrid() *Grid {
	return this.grid
}

func (this *User) setGrid(grid *Grid) {
	this.grid = grid
}

func (this *User) getRoom() *Room {
	return this.room
}

func (this *User) getImage() *UserImage {
	image := UserImage{}
	image.I = this.getIndex()
	image.UserID = common.ParseInt(this.getUserID())
	image.GridIndex = this.getGrid().getIndex()
	return &image
}

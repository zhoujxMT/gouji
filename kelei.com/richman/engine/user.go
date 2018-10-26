package engine

import (
	"net"

	"kelei.com/utils/common"
)

type User struct {
	index  int
	userid string
	conn   net.Conn
	room   *Room
	grid   *Grid
}

type UserImage struct {
	I         int
	UserID    int
	GridIndex int
}

func NewUser() *User {
	user := User{}
	return &user
}

/*
	逻辑实现
*/

func (this *User) enterRoom(room *Room) {
	this.room = room
	room.addUser(this)
}

/*
	玩家是否在房间中
*/
func (this *User) inRoom() bool {
	return this.getRoom() != nil
}

/*
	退出房间
*/
func (this *User) exitRoom() {
	this.room.removeUser(this)
	this.room = nil
}

/*
	掷骰子
*/
func (this *User) dicing() int {
	step := common.Random(1, 6)
	//玩家移动
	this.move(step)
	return step
}

/*
	玩家移动
*/
func (this *User) move(step int) {
	grid := this.getGrid()
	nextOneGrid := grid.getNextOneGrid(step)
	this.leaveGrid()
	this.enterGrid(nextOneGrid)
	this.pushGrid([]*Grid{grid, nextOneGrid})
}

/*
	离开网格
*/
func (this *User) leaveGrid() {
	grid := this.getGrid()
	grid.userLeave()
	this.setGrid(nil)
}

/*
	进入网格
*/
func (this *User) enterGrid(grid *Grid) {
	grid.userEnter(this)
	this.setGrid(grid)
}

/*
	吃道具
*/
func (this *User) eatItem(item *Item) {

}

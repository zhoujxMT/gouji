package engine

import (
	"kelei.com/utils/logger"
)

const (
	STATUS_SETOUT = iota
	STATUS_MATCHING
)

type Room struct {
	roomid         string
	pcount         int
	status         int
	users          []*User
	gridSystem     *GridSystem
	roomUserSystem *RoomUserSystem
	itemSystem     *ItemSystem
	turretSystem   *TurretSystem
}

func NewRoom() *Room {
	room := &Room{}
	room.gridSystem = NewGridSystem(room)
	room.roomUserSystem = NewRoomUserSystem(room)
	room.itemSystem = NewItemSystem(room)
	room.turretSystem = NewTurretSystem(room)
	room.pcount = 2
	return room
}

func (this *Room) addUser(user *User) {
	user.setIndex(len(this.users))
	this.users = append(this.users, user)
	//人数已满
	if len(this.GetUsers()) >= this.GetPCount() {
		//开赛
		this.start()
	}
}

func (this *Room) removeUser(user *User) {
	for i, u := range this.GetUsers() {
		if u == user {
			this.users = append(this.users[:i], this.users[i+1:]...)
			break
		}
	}
}

func (this *Room) push(funcName string, content *string) {
	users := this.GetUsers()
	for _, user := range users {
		user.push(funcName, content)
	}
}

/*
	=========================================逻辑实现
*/

/*
	人数已满,开赛
*/
func (this *Room) start() {
	logger.Debugf("开赛")
	this.setStatus(STATUS_MATCHING)
	this.build()
	this.pushAll()
}

/*
	构建
*/
func (this *Room) build() {
	this.roomUserSystem.fill()
	this.gridSystem.fill()
	this.itemSystem.fill()
	this.turretSystem.fill()
}

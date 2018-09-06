package engine

import (
	"fmt"
)

const (
	STATUS_SETOUT = iota
	STATUS_MATCHING
)

type Room struct {
	roomid     string
	gridSystem *GridSystem
	users      []*User
	pcount     int
	status     int
}

func NewRoom() *Room {
	room := Room{}
	room.gridSystem = NewGridSystem()
	room.pcount = 2
	return &room
}

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

func (this *Room) GetPCount() int {
	return this.pcount
}

func (this *Room) ISMatching() bool {
	return this.GetStatus() == STATUS_MATCHING
}

func (this *Room) ISSetout() bool {
	return this.GetStatus() == STATUS_SETOUT
}

func (this *Room) addUser(user *User) {
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

func (this *Room) GetUsers() []*User {
	return this.users
}

func (this *Room) push(content string) {
	users := this.GetUsers()
	for _, user := range users {
		user.push(content)
	}
}

func (this *Room) start() {
	fmt.Println("开赛")
	this.setStatus(STATUS_MATCHING)
	this.push("i love you")
}

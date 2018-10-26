package engine

import (
	"encoding/json"

	"kelei.com/utils/logger"
)

type RoomUserSystem struct {
	userCount int
	users     []*User
	room      *Room
}

type RoomUserSystemImage struct {
	Users []UserImage
}

func NewRoomUserSystem(room *Room) *RoomUserSystem {
	roomUserSystem := RoomUserSystem{}
	roomUserSystem.userCount = 2
	roomUserSystem.users = []*User{}
	roomUserSystem.room = room
	roomUserSystem.build()
	return &roomUserSystem
}

func (this *RoomUserSystem) build() {
}

func (this *RoomUserSystem) AddUser(userid string) *User {
	user := NewUser()
	user.setUserID(userid)
	this.users = append(this.users, user)
	return user
}

func (this *RoomUserSystem) removeUser(user *User) {
	//	this.users=append(this.users[])
}

func (this *RoomUserSystem) GetUserCount() int {
	return len(this.users)
}

func (this *RoomUserSystem) GetUser(index int) *User {
	user := this.users[index]
	return user
}

func (this *RoomUserSystem) getRoom() *Room {
	return this.room
}

func (this *RoomUserSystem) GetUsers() []*User {
	return this.users
}

func (this *RoomUserSystem) setUsers(users []*User) {
	this.users = users
}

func (this *RoomUserSystem) fill() {
	room := this.getRoom()
	users := room.GetUsers()
	this.setUsers(users)
}

func (this *RoomUserSystem) getImage() *string {
	image := RoomUserSystemImage{[]UserImage{}}
	users := this.GetUsers()
	for _, user := range users {
		userImage := *user.getImage()
		image.Users = append(image.Users, userImage)
	}
	b, err := json.Marshal(image)
	logger.CheckFatal(err)
	msg := string(b)
	return &msg
}

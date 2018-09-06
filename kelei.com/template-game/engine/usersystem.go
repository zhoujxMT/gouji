package engine

import (
	"net"
)

type UserSystem struct {
	users map[string]*User
}

func NewUserSystem() *UserSystem {
	userSystem := UserSystem{}
	userSystem.users = make(map[string]*User)
	return &userSystem
}

func (this *UserSystem) AddUser(userid string, conn net.Conn) *User {
	user := this.users[userid]
	if user == nil {
		user = NewUser()
		user.setConn(conn)
		user.setUserID(userid)
		this.users[userid] = user
	}
	return user
}

func (this *UserSystem) removeUser(user *User) {
	delete(this.users, user.getUserID())
}

func (this *UserSystem) GetUserCount() int {
	return len(this.users)
}

func (this *UserSystem) GetUser(userid string) *User {
	user := this.users[userid]
	return user
}

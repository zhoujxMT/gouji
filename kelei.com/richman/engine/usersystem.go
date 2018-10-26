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
	} else {
		user.setConn(conn)
	}
	return user
}

func (this *UserSystem) Clear() {
	this.users = make(map[string]*User)
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

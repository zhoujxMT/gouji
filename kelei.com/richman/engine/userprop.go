package engine

import (
	"net"
)

func (this *User) getUserID() string {
	return this.userid
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

func (this *User) getRoom() *Room {
	return this.room
}

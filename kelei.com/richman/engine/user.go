package engine

import (
	"fmt"
	"net"
	"time"

	"kelei.com/utils/frame"
	"kelei.com/utils/logger"
)

type User struct {
	userid string
	conn   net.Conn
	room   *Room
}

func NewUser() *User {
	user := User{}
	return &user
}

func (this *User) push(content string) {
	go func() {
		time.Sleep(time.Millisecond)
		conn := this.getConn()
		xServer := frame.GetRpcxServer()
		content = fmt.Sprintf("%s:%s", this.getUserID(), content)
		err := xServer.SendMessage(conn, "service_path", "service_method", nil, []byte(content))
		logger.CheckError(err, fmt.Sprintf("failed to send messsage to %s : ", conn.RemoteAddr().String()))
	}()
}

/*
	逻辑实现
*/

func (this *User) enterRoom(room *Room) {
	this.room = room
	room.addUser(this)
}

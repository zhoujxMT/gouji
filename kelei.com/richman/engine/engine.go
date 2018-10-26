package engine

import (
	"net"
	"strconv"

	"kelei.com/utils/common"
	"kelei.com/utils/logger"
	"kelei.com/utils/rpcs"
)

type Engine struct {
	userSystem *UserSystem
	roomSystem *RoomSystem
}

func NewEngine() *Engine {
	logger.Infof("[启动引擎]")
	engine := Engine{}
	engine.userSystem = NewUserSystem()
	engine.roomSystem = NewRoomSystem()
	return &engine
}

func (this *Engine) GetRoomSystem() *RoomSystem {
	return this.roomSystem
}

func (this *Engine) GetUserSystem() *UserSystem {
	return this.userSystem
}

func (this *Engine) getUserID(args *rpcs.Args) string {
	p := args.V["args"].([]interface{})
	userid := p[0].(string)
	return userid
}

func (this *Engine) getUser(args *rpcs.Args) *User {
	userid := this.getUserID(args)
	user := this.userSystem.GetUser(userid)
	return user
}

func (this *Engine) ClientConnClose(args *rpcs.Args) *rpcs.Reply {
	user := this.getUser(args)
	room := user.getRoom()
	if room.ISSetout() {
		room.removeUser(user)
		this.userSystem.removeUser(user)
	}
	return &rpcs.Reply{nil, common.SC_OK}
}

func (this *Engine) Connect(args *rpcs.Args) *rpcs.Reply {
	userid := this.getUserID(args)
	//	this.temporary(args)
	this.userSystem.AddUser(userid, args.V["clientConn"].(net.Conn))
	return &rpcs.Reply{nil, common.SC_OK}
}

func (this *Engine) temporary(args *rpcs.Args) {
	user := this.getUser(args)
	if user != nil {
		this.userSystem.removeUser(user)
		if user.getRoom() != nil {
			user.exitRoom()
		}
	}
}

/*
	匹配房间
	push:所有模块的镜像
*/
func (this *Engine) Matching(args *rpcs.Args) *rpcs.Reply {
	user := this.getUser(args)
	room := user.getRoom()
	if room != nil {
		room.pushAll()
		return &rpcs.Reply{nil, common.SC_OK}
	}
	room = this.roomSystem.matchingRoom()
	user.enterRoom(room)
	return &rpcs.Reply{nil, common.SC_OK}
}

/*
	重置引擎
*/
func (this *Engine) Reset(args *rpcs.Args) *rpcs.Reply {
	this.GetRoomSystem().Clear()
	this.GetUserSystem().Clear()
	return &rpcs.Reply{nil, common.SC_OK}
}

/*
	掷骰子
	out:点数
*/
func (this *Engine) Dicing(args *rpcs.Args) *rpcs.Reply {
	user := this.getUser(args)
	if !user.inRoom() {
		return &rpcs.Reply{nil, common.SC_ERR}
	}
	if user.getRoom().ISSetout() {
		return &rpcs.Reply{nil, common.SC_ERR}
	}
	stepCount := strconv.Itoa(user.dicing())
	return &rpcs.Reply{&stepCount, common.SC_OK}
}

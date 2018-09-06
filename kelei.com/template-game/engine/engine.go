package engine

import (
	"net"

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

func (this *Engine) ClientConnClose(args *rpcs.Args) *string {
	res := common.STATUS_CODE_SUCCEED
	user := this.getUser(args)
	room := user.getRoom()
	if room.ISSetout() {
		room.removeUser(user)
		this.userSystem.removeUser(user)
	}
	return &res
}

func (this *Engine) Connect(args *rpcs.Args) *string {
	res := common.STATUS_CODE_SUCCEED
	userid := this.getUserID(args)
	this.userSystem.AddUser(userid, args.V["clientConn"].(net.Conn))
	return &res
}

func (this *Engine) Matching(args *rpcs.Args) *string {
	res := common.STATUS_CODE_SUCCEED
	user := this.getUser(args)
	if user.getRoom() != nil {
		return &res
	}
	room := this.roomSystem.matchingRoom()
	user.enterRoom(room)
	return &res
}

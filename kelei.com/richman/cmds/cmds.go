package cmds

import (
	"bytes"
	"fmt"

	eng "kelei.com/richman/engine"
	"kelei.com/utils/frame/command"
	"kelei.com/utils/logger"
)

var (
	engine *eng.Engine
)

func Inject(engine_ *eng.Engine) {
	engine = engine_
}

func GetCmds() map[string]func(string) {
	var commands = map[string]func(string){}
	//begin 自定义方法
	commands["test"] = test
	commands["usercount"] = usercount
	commands["roomcount"] = roomcount
	commands["roominfo"] = roominfo
	//end
	command.CreateHelp(commands)
	return commands
}

func test(cmdVal string) {
	logger.Infof("test succees")
}

func usercount(cmdVal string) {
	logger.Infof("玩家数量:%d", engine.GetUserSystem().GetUserCount())
}

func roomcount(cmdVal string) {
	logger.Infof("房间数量:%d", engine.GetRoomSystem().GetRoomCount())
}

func roominfo(cmdVal string) {
	rooms := engine.GetRoomSystem().GetRooms()
	buff := bytes.Buffer{}
	buff.WriteString("{\n")
	for _, room := range rooms {
		buff.WriteString(fmt.Sprintf("     id:%s , 房间人数:%d , 房间状态:%d\n", room.GetRoomID(), len(room.GetUsers()), room.GetStatus()))
	}
	buff.WriteString("}")
	logger.Infof(buff.String())
	logger.Infof("房间数量:%d", engine.GetRoomSystem().GetRoomCount())
}

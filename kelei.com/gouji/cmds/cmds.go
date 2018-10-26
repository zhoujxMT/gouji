package cmds

import (
	"bytes"
	"fmt"

	eng "kelei.com/gouji/engine"
	"kelei.com/utils/common"
	"kelei.com/utils/frame/command"
	"kelei.com/utils/frame/config"
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
	commands["usercount"] = usercount
	commands["roomcount"] = roomcount
	commands["roominfo"] = roominfo
	commands["pcount"] = pcount
	commands["cardcount"] = cardcount
	commands["geming"] = geming
	commands["tuoguan"] = tuoguan
	//end
	commands["mode"] = config.Mode
	command.CreateHelp(commands)
	return commands
}

func usercount(cmdVal string) {
	logger.Infof("玩家数量:%d", eng.UserManage.GetUserCount())
}

func roomcount(cmdVal string) {
	logger.Infof("房间数量:%d", eng.RoomManage.GetRoomCount())
}

func roominfo(cmdVal string) {
	rooms := eng.RoomManage.GetRooms()
	buff := bytes.Buffer{}
	buff.WriteString("{\n")
	for _, room := range rooms {
		buff.WriteString(fmt.Sprintf("     id:%s,比赛类型:%d,房间类型:%d,房间人数:%d,观战人数:%d\n", *room.GetRoomID(), room.GetMatchID(), room.GetRoomType(), room.GetPCount(), room.GetIdlePCount()))
	}
	buff.WriteString("}")
	logger.Infof(buff.String())
	logger.Infof("房间数量:%d", eng.RoomManage.GetRoomCount())
}

func pcount(commVal string) {
	if commVal == "" {
		logger.Infof("开局人数 : %d", eng.GetPCount())
	} else {
		if pcount, err := common.CheckNum(commVal); err == nil {
			eng.SetPCount(pcount)
		}
	}
}

func cardcount(commVal string) {
	if commVal == "" {
		logger.Infof("每个人的牌数量 : %d", eng.GetCardCount())
	} else {
		if cardcount, err := common.CheckNum(commVal); err == nil {
			eng.SetCardCount(cardcount)
		}
	}
}

func geming(commVal string) {
	if commVal == "" {
		logger.Infof("革命的人数 : %d", eng.GetRevolutionCount())
	} else {
		if count, err := common.CheckNum(commVal); err == nil {
			eng.SetRevolutionCount(count)
		}
	}
}

func tuoguan(commVal string) {
	if commVal == "" {
		logger.Infof("请设置托管参数")
	} else {
		if v, err := common.CheckNum(commVal); err == nil {
			rooms := eng.RoomManage.GetRooms()
			for _, room := range rooms {
				room.SetAllUsersTrusteeshipStatus(v == 1)
			}
		}
	}
}

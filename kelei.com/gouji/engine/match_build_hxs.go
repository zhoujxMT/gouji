package engine

import (
	"strconv"

	"kelei.com/utils/common"
	"kelei.com/utils/logger"
)

/*
进入好友同玩的房间
in:roomid
out:-1比赛房间不存在
	1成功进入
push:1.不在比赛中,推送房间状态信息
	 2.在比赛中,推送比赛残局
*/
func EnterHXS(args []string) (*string, string, *User) {
	res := common.Res_Succeed
	userid := args[0]
	matchID, err := strconv.Atoi(args[1])
	logger.CheckFatal(err, "Matching:1")
	roomType_s := args[2]
	currentRoomID := args[4]
	user := UserManage.GetUser(&userid)
	//进入房间的判断
	res = *checkEnterRoom(userid, matchID, roomType_s)
	//不是成功 && 不是残局
	if res != "1" && res != "2" {
		return &res, currentRoomID, nil
	}
	roomType, err3 := strconv.Atoi(roomType_s)
	logger.CheckFatal(err3, "Matching:2")
	if currentRoomID == "-1" { //不在比赛中
		if user.getRoom() == nil {
			user.reset()
			//获取人数不满的房间
			room := getVacancyRoom(matchID, roomType)
			//进入此房间
			user.enterRoom(room)
			//坐下
			user.sitDownAuto()
			//准备
			Setout(args)
		}
	}
	return &res, currentRoomID, user
}

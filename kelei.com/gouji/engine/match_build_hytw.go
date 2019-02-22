/*
比赛构建-好友同玩
*/

package engine

import (
	"strconv"

	. "kelei.com/utils/common"
	"kelei.com/utils/logger"
)

const (
	SeatStatus_StandUp = iota
	SeatStatus_SitDown
)

/*
进入好友同玩的房间
in:roomid
out:-1比赛房间不存在
	1成功进入
push:1.不在比赛中,推送房间状态信息
	 2.在比赛中,推送比赛残局
*/
func EnterHYTW(args []string) (*string, string, *User) {
	res := "1"
	userid := args[0]
	currentRoomID := args[3]
	roomid := args[2]
	user := UserManage.GetUser(&userid)
	if currentRoomID == "-1" { //不在比赛中
		if user.getRoom() == nil {
			room := RoomManage.GetRoom(roomid)
			//房间不存在
			if room == nil {
				res = "-1"
				return &res, currentRoomID, nil
			}
			user.reset()
			user.enterRoom(room)
			user.setStatus(UserStatus_NoSitDown)
			//推送房间的状态信息
			room.matchingPush(nil)
		}
	}
	return &res, currentRoomID, user
}

/*
站起坐下
in:状态(0站起1坐下),坐位编号
out:-1坐位编号错误 -2坐位已有人 -3开赛之后不允许站起 -4重复操作
	1成功
push:推送房间状态信息
*/
func SitDown(args []string) *string {
	res := "1"
	userid := args[0]
	status, err := strconv.Atoi(args[1])
	logger.CheckFatal(err)
	user := UserManage.GetUser(&userid)
	if user == nil {
		res := Res_Unknown
		return &res
	}
	room := user.getRoom()
	//开赛之后,不允许站起坐下
	if room.isMatching() {
		res = "-3"
		return &res
	}
	//重复操作直接返回
	if user.getStatus() == UserStatus_NoSitDown { //在站起的状态下,站起
		if status == SeatStatus_StandUp {
			res = "-4"
			return &res
		}
	} else { //在坐下的状态下,坐下
		if status == SeatStatus_SitDown {
			seatIndex, err := strconv.Atoi(args[2])
			logger.CheckFatal(err)
			//如果坐的是同一个位置
			if user.getIndex() == seatIndex {
				res = "-4"
				return &res
			}
		}
	}
	if status == SeatStatus_SitDown {
		seatIndex, err := strconv.Atoi(args[2])
		logger.CheckFatal(err)
		//坐位编号错误
		if !(seatIndex >= 0 && seatIndex <= pcount-1) {
			res = "-1"
			return &res
		}
		//坐位已有人
		if room.getUsers()[seatIndex] != nil {
			res = "-2"
			return &res
		}
		//坐的是其它位置,先从当前位置站起,再坐入新的位置
		user.standUp()
		user.sitDown(seatIndex)
	} else {
		user.standUp()
	}
	//推送房间的状态信息
	room.matchingPush(nil)
	return &res
}

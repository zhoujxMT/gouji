/*
比赛构建
*/

package engine

import (
	"strconv"

	"github.com/garyburd/redigo/redis"

	. "kelei.com/utils/common"
	"kelei.com/utils/logger"
)

/*
匹配房间
in:比赛类型,X,匹配方式
out:
	【够级英雄 & 好友同玩】
		-2比赛类型未开放,-3房间类型未开放,-4元宝数不足,-5等级不够,-6房间不存在
		1匹配成功 2残局
	【海选赛】
		-2比赛类型未开放
		1匹配成功 2残局
des:比赛类型(0够级英雄1好友同玩2海选赛)
	X(够级英雄:roomType,好友同玩:roomid,海选赛:0)
	匹配方式(0匹配1匹配>进入>准备2匹配>进入)
*/
func Matching(args []string) *string {
	userid := args[0]
	currentRoomID := args[4]
	if currentRoomID != "-1" {
		res := "2"
		return &res
	}
	matchID, err := strconv.Atoi(args[1])
	logger.CheckFatal(err, "Matching:1")
	x := args[2]
	mode_s := "0"
	if len(args) >= 5 {
		mode_s = args[3]
	}
	mode, err2 := strconv.Atoi(mode_s)
	logger.CheckFatal(err2, "Matching:2")
	res := Res_Succeed
	if mode == MatchingMode_Normal {
		res = *checkEnterRoom(userid, matchID, x)
		res_int, _ := strconv.Atoi(res)
		//海选赛只需要判断是否开放就可进入
		if matchID == Match_HXS && res_int < -2 {
			res = "1"
		}
	} else if mode == MatchingMode_Enter {
		res = *EnterRoom(args)
	} else if mode == MatchingMode_EnterAndSetout {
		res = *EnterRoom(args)
		res_i, _ := strconv.Atoi(res)
		if res_i > 0 {
			res = *Setout(args)
		}
	}
	return &res
}

/*
进入房间
in:比赛类型,X
out:
	【够级英雄 & 好友同玩】
		-2比赛类型未开放,-3房间类型未开放,-4元宝数不足,-5等级不够,-6房间不存在
		1匹配成功 2残局
	【海选赛】
		-2比赛类型未开放,-4元宝数不足,-5等级不够,-6比赛次数不足,-7海选赛已截止
		1匹配成功 2残局
des:比赛类型(0够级英雄1好友同玩2海选赛)
	X(够级英雄:roomType,好友同玩:roomid,海选赛:0)
*/
func EnterRoom(args []string) *string {
	var res *string
	userid := args[0]
	user := UserManage.GetUser(&userid)
	matchID, err := strconv.Atoi(args[1])
	logger.CheckFatal(err, "EnterRoom")
	currentRoomID := ""
	if matchID == Match_GJYX {
		res, currentRoomID, user = EnterGJYX(args)
	} else if matchID == Match_HYTW {
		res, currentRoomID, user = EnterHYTW(args)
	} else if matchID == Match_HXS {
		res, currentRoomID, user = EnterHXS(args)
	}
	//在房间中
	if currentRoomID != "-1" {
		//比赛结束后,检测不在线的玩家都被从UserManager中删除了,所以user变成了空
		if user == nil {
			r := "-6"
			res = &r
		} else {
			//设置玩家在线了
			user.setOnline(true)
			user.pushEndGame()
		}
	}
	user = UserManage.GetUser(&userid)
	if user != nil && user.getRoom() != nil {
		res = user.getRoom().GetRoomID()
	}
	return res
}

//快速匹配获取房间类型(够级英雄)
func getRoomTypeWithFastMatching() int {
	return 1
}

/*
换桌
in:
out:-1比赛中不能换桌
	1换桌成功
push:离开房间的状态推送、新进入房间的状态推送
*/
func ChangeTable(args []string) *string {
	res := Res_Succeed
	userid := args[0]
	user := UserManage.GetUser(&userid)
	room := user.getRoom()
	matchID := room.GetMatchID()
	roomType := room.GetRoomType()
	//比赛中不能换桌
	if user.isMatching() {
		res = "-1"
		return &res
	}
	//退出当前房间
	user.exitRoom()
	//进入新房间
	EnterGJYX([]string{userid, strconv.Itoa(matchID), strconv.Itoa(roomType), strconv.Itoa(MatchingMode_Enter), "-1"})
	res = *user.getRoom().GetRoomID()
	return &res
}

/*
客户端关闭连接
in:
out:无返回值
*/
func ClientConnClose(args []string) *string {
	res := Res_NoBack
	userid := args[0]
	if closeJudgment(&userid) {
		return &res
	}
	closeMatchUser(&userid)
	closeKeepAliveUser(&userid)
	return &res
}

//裁判关闭连接的方法
func closeJudgment(userid *string) bool {
	user := UserManage.GetUser(userid)
	if user == nil {
		return false
	}
	judgmentUser := user.getRoom().getJudgmentUser()
	//不是裁判关闭连接
	if judgmentUser != user {
		return false
	}
	//裁判离场
	if judgmentUser.getUserType() == TYPE_JUDGMENT {
		logger.Debugf("裁判关闭连接的方法")
		room := judgmentUser.getRoom()
		room.pause()
		return true
	}
	return false
}

//删除比赛中的玩家
func closeMatchUser(userid *string) {
	user := UserManage.GetUser(userid)
	if user == nil {
		return
	}
	user.close()
}

//删除长连接的玩家
func closeKeepAliveUser(userid *string) {
	user := UserManage.GetUserFromKeepAliveUsers(userid)
	if user == nil {
		return
	}
	user.closeFromKeepAliveUsers()
}

/*
退出比赛
in:
out:无返回值
push:1.在比赛中,推送给所有人有人逃跑
	 2.不在比赛中,推送房间状态信息

*/
func ExitMatch(args []string) *string {
	res := Res_Unknown
	userid := args[0]
	user := UserManage.GetUser(&userid)
	if user == nil {
		return &res
	}
	if user.isMatching() {
		user.exitMatchRoom()
	} else {
		user.exitRoom()
	}
	return &res
}

//查找人数不满的房间
func getVacancyRoom(matchID int, roomType int) *Room {
	rooms := RoomManage.GetRooms()
	vacancyRooms := []*Room{}
	for _, room := range rooms {
		if room.GetMatchID() == matchID && room.GetRoomType() == roomType && room.GetPCount() < pcount && !room.isMatching() {
			vacancyRooms = append(vacancyRooms, room)
		}
	}
	if len(vacancyRooms) <= 0 {
		room := RoomManage.AddRoom()
		room.setMatchID(matchID)
		room.setRoomType(roomType)
		room.setRoomBaseInfo()
		vacancyRooms = append(vacancyRooms, room)
	}
	return vacancyRooms[Random(0, len(vacancyRooms))]
}

/*
准备
in:
out:-1元宝数不足
	1准备成功
push:1.推送房间状态信息
	 2.推送开局信息
*/
func Setout(args []string) *string {
	userid := &args[0]
	user := UserManage.GetUser(userid)
	if user == nil {
		res := Res_Unknown
		return &res
	}
	room := user.getRoom()
	if user == nil {
		res := Res_Unknown
		logger.Errorf("Setout 玩家没在房间中")
		return &res
	}
	roomData, _ := redis.Ints(room.GetRoomData("PlayIngot"))
	playIngot := roomData[0]
	//获取玩家元宝数
	userIngot := user.getIngot()
	if userIngot < playIngot {
		res := "-1"
		return &res
	}
	//只有没准备的情况下，才可以准备
	if user.getStatus() == UserStatus_NoSetout {
		user.setStatus(UserStatus_Setout)
		user.close_countDown_setOut()
		room := user.getRoom()
		if room != nil {
			if room.GetMatchID() == Match_HXS {
				//推送海选赛匹配信息(没进入房间)
				room.matchingHXSPush(nil)
			} else {
				//推送房间的状态信息(已进入房间)
				room.matchingPush(nil)
			}
			//判断玩家准备的情况,决定是否开局
			count := room.getSetoutCount()
			if count >= pcount {
				//开局
				room.opening()
			}
		}
	}
	res := "1"
	return &res
}

/*
获取座位状态的推送
in:
out:1成功,其它失败
push:Matching_Push,64$1$0$5$0|||||#等待席#当前轮次
*/
func GetMatchingPush(args []string) *string {
	userid := &args[0]
	user := UserManage.GetUser(userid)
	if user == nil {
		return &Res_Unknown
	}
	room := user.getRoom()
	if room == nil {
		return &Res_Unknown
	}
	//推送房间的状态信息
	room.matchingPush(user)
	return &Res_Succeed
}

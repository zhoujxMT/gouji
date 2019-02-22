/*
比赛构建-够级英雄
*/

package engine

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	. "kelei.com/utils/common"
	"kelei.com/utils/logger"
)

/*
进入够级英雄
in:比赛类型,房间类型
out:-2比赛类型未开放,-3房间类型未开放,-4元宝数不足,-5等级不够
	是否成功进入,
push:1.不在比赛中,推送房间状态信息
	 2.在比赛中,推送比赛残局
*/
func EnterGJYX(args []string) (*string, string, *User) {
	res := Res_Succeed
	userid := args[0]
	matchID, err := strconv.Atoi(args[1])
	logger.CheckFatal(err, "Matching:1")
	roomType_s := args[2]
	currentRoomID := args[4]
	user := UserManage.GetUser(&userid)
	//进入房间的判断
	res = *checkEnterRoom(userid, matchID, roomType_s)
	//进入成功 && 残局
	if res != "1" && res != "2" {
		return &res, currentRoomID, nil
	}
	roomType, err3 := strconv.Atoi(roomType_s)
	logger.CheckFatal(err3, "Matching:2")
	//不在比赛中
	if currentRoomID == "-1" {
		if user.getRoom() == nil {
			user.reset()
			room := &Room{}
			userIndex := 0
			gameRule := room.GetGameRuleConfig()
			if gameRule == GameRule_Normal {
				//获取人数不满的房间
				room = getVacancyRoom(matchID, roomType)
			} else if gameRule == GameRule_Record {
				uid := *user.getUID()
				reg := regexp.MustCompile(`[0-9]+`)
				uid_str := reg.FindAllString(uid, -1)
				uid_int, _ := strconv.Atoi(strings.Join(uid_str, ""))
				roomid := 1
				if uid_int <= 16 {
					roomid = 1
				} else if uid_int <= 26 {
					roomid = 2
				} else if uid_int <= 36 {
					roomid = 3
				}
				userIndex = uid_int % 10
				room = RoomManage.GetRoom(strconv.Itoa(roomid))
			}
			//进入此房间
			user.enterRoom(room)
			//坐下
			if gameRule == GameRule_Record {
				user.sitDownPigeon(userIndex - 1)
			} else {
				user.sitDownAuto()
			}
			//准备倒计时
			user.countDown_setOut(time.Second * 10)
			//推送房间的状态信息
			room.matchingPush(nil)
		}
	}
	return &res, currentRoomID, user
}

/*
根据玩家元宝获取快速匹配的房间类型(够级英雄)
in:
out:roomtype
*/
func GetFMRoomType(args []string) *string {
	userid := args[0]
	//创建玩家
	user := UserManage.createUser()
	user.setUserID(&userid)
	ingot := user.getIngot()
	userLevel := user.getLevel()
	roomType := RoomType_Primary
	room := Room{}
	allRoomData := *room.getAllRoomData()
	arrRoomData := strings.Split(allRoomData, "|")
	for _, roomData := range arrRoomData {
		arrRoomData := strings.Split(roomData, "$")
		roomtype_, _ := strconv.Atoi(arrRoomData[0])
		matchingIngot, _ := strconv.Atoi(arrRoomData[3])
		levelLimit, _ := strconv.Atoi(arrRoomData[5])
		if matchingIngot != 0 {
			//等级限制,金币限制
			if userLevel >= levelLimit && ingot >= matchingIngot {
				roomType = roomtype_
			}
		}
	}
	roomtype_s := strconv.Itoa(roomType)
	return &roomtype_s
}

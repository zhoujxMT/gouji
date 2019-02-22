/*
游戏服务器操作-房间
*/

package engine

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/garyburd/redigo/redis"

	. "kelei.com/utils/common"
	"kelei.com/utils/frame"
	"kelei.com/utils/json"
	"kelei.com/utils/logger"
	"kelei.com/utils/rpcs"
)

/*
获取比赛类型下的房间类型信息
in:比赛类型
out:-2没有房间类型信息
	房间类型$进场要求$底分$人数|
*/
func GetMatchRoomInfo(args []string) *string {
	res := Res_Unknown
	matchID, err := strconv.Atoi(args[1])
	logger.CheckError(err, "GetMatchRoomInfo:1")
	if matchID == Match_GJYX {
		res = *getGJYXRoomInfo()
	}
	return &res
}

/*
获取"青岛够级"的房间类型信息
in:
out:-2没有房间类型信息
	房间类型$进场要求$底分$人数|
*/
func getGJYXRoomInfo() *string {
	res := "-2"
	buff := bytes.Buffer{}
	room := Room{}
	allRoomData := *room.getAllRoomData()
	arrAllRoomData := strings.Split(allRoomData, "|")
	arr := roompcountT[Match_GJYX]
	for _, roomData := range arrAllRoomData {
		arrRoomData_s := strings.Split(roomData, "$")
		arrRoomData := StrArrToIntArr(arrRoomData_s)
		roomType, enterIngot, multiple := arrRoomData[0], arrRoomData[1], arrRoomData[2]
		pcount := arr[roomType]
		buff.WriteString(fmt.Sprintf("%d$%d$%d$%d|", roomType, enterIngot, multiple, pcount))
	}
	res = RemoveLastChar(buff)
	return &res
}

/*
是否可进入房间
in:比赛类型,房间类型
out:-2比赛类型未开放,-3房间类型不存在,-4元宝数不足,-5等级不够
	1可以进入房间
*/
func checkEnterRoom(userid string, matchID int, x string) *string {
	res := "-2"
	matchInfo, err2 := redis.Ints(GetMatchInfo(matchID, "status"))
	logger.CheckFatal(err2, "checkEnterRoom:1")
	status := matchInfo[0]
	//比赛类型未开放
	if status == 0 {
		return &res
	}
	if matchID == Match_GJYX {
		roomType, err := strconv.Atoi(x)
		logger.CheckFatal(err, "checkEnterRoom:2")
		res = checkEnterGJYX(userid, roomType)
	} else if matchID == Match_HYTW {
		roomID := x
		res = checkEnterHYTW(roomID)
	} else if matchID == Match_KDFS {
	} else if matchID == Match_HXS {
		res = checkEnterHXS(userid)
	}
	return &res
}

/*
进入好友同玩的房间
in:roomid
out:-6房间不存在
	1可以进入房间
*/
func checkEnterHYTW(roomID string) string {
	res := "1"
	room := RoomManage.GetRoom(roomID)
	//房间不存在
	if room == nil {
		res = "-6"
		return res
	}
	return res
}

//获取房间类型是否存在
func roomTypeExists(roomType int) bool {
	if roomType >= RoomType_Primary && roomType <= RoomType_Tribute {
		return true
	}
	return false
}

/*
进入青岛够级的房间
in:房间类型
out:-3房间类型不存在,-4元宝数不足,-5等级不够
	1可以进入房间
*/
func checkEnterGJYX(userid string, roomType int) string {
	res := "1"
	//创建玩家
	user := UserManage.createUser()
	user.setUserID(&userid)
	if !roomTypeExists(roomType) {
		res = "-3"
		return res
	}
	//创建房间
	room := Room{}
	room.setRoomType(roomType)
	//获取进场要求元宝数和等级
	roomData, err := redis.Ints(room.GetRoomData("enterIngot", "level"))
	logger.CheckFatal(err, "enterGJYX")
	enterIngot := roomData[0]
	level := roomData[1]
	//获取玩家元宝数
	userIngot := user.getIngot()
	if userIngot < enterIngot {
		res = "-4"
		return res
	}
	if user.getLevel() < level {
		res = "-5"
		return res
	}
	return res
}

/*
进入海选赛的房间
in:房间类型
out:-2比赛类型未开放,-4元宝数不足,-5等级不够,-6比赛次数不足,-7海选赛已截止
	1可以进入房间
*/
func checkEnterHXS(userid string) string {
	res := "1"
	roomType := 1
	//创建玩家
	user := UserManage.GetUser(&userid)
	//创建房间
	room := Room{}
	room.setRoomType(roomType)
	//获取进场要求元宝数和等级
	roomDataAudition, err := redis.Ints(room.GetRoomDataAudition("enterIngot", "level"))
	logger.CheckFatal(err, "enterHXS")
	enterIngot := roomDataAudition[0]
	level := roomDataAudition[1]
	//获取玩家元宝数
	userIngot := user.getIngot()
	if userIngot < enterIngot {
		res = "-4"
		return res
	}
	if user.getLevel() < level {
		res = "-5"
		return res
	}
	err2 := db.QueryRow("call AuditionUserEnter(?)", *user.getUserID()).Scan(&res)
	logger.CheckFatal(err2, "AuditionUserEnter")
	return res
}

//好友同玩
type HYTW struct {
	Ingot int
}

//获得好友同玩的配置信息
func getHYTWData() HYTW {
	room := Room{}
	room.setMatchID(Match_HYTW)
	content := *room.GetMatchData()
	hytw := HYTW{}
	js := json.NewJsonStruct()
	js.LoadByString(content, &hytw)
	return hytw
}

/*
获取创建好友同玩房间所需元宝数
in:
out:元宝数量|
*/
func GetHYTWIngot(args []string) *string {
	hytw := getHYTWData()
	res := strconv.Itoa(hytw.Ingot)
	res = fmt.Sprintf("%s|%s|%s", res, res, res)
	return &res
}

/*
创建房间
in:matchid|*
out:*
*/
func CreateRoom(args []string) *string {
	res := Res_Unknown
	matchid, err := strconv.Atoi(args[1])
	logger.CheckFatal(err)
	if matchid == Match_HYTW {
		res = CreateHYTWRoom(args)
	}
	return &res
}

/*
创建好友同玩房间
in:matchid|模式(0够级英雄 1开点发四)|进贡|革命|一轮局数
out:-1模式不存在,-2轮数不符合,-3元宝不足
	roomid
*/
func CreateHYTWRoom(args []string) string {
	res := Res_Unknown
	userid := args[0]
	var mode, tribute, revolution, count int
	var err error
	mode, err = strconv.Atoi(args[2])
	logger.CheckFatal(err)
	tribute, err = strconv.Atoi(args[3])
	logger.CheckFatal(err)
	revolution, err = strconv.Atoi(args[4])
	logger.CheckFatal(err)
	count, err = strconv.Atoi(args[5])
	logger.CheckFatal(err)
	//模式不存在
	if !(mode >= 0 && mode <= 1) {
		return "-1"
	}
	//轮数不符合
	if !(count == 3 || count == 5 || count == 7) {
		return "-2"
	}
	user := User{}
	user.setUserID(&userid)
	ingot := user.getIngot()
	revolution = revolution
	hytw := getHYTWData()
	if ingot < hytw.Ingot {
		return "-3"
	}
	user.updateIngot(-hytw.Ingot, 1)
	//生成一个房间
	room := RoomManage.AddRoom()
	room.setMatchID(Match_HYTW)
	room.setInnings(count)
	room.setTribute(tribute == 1)
	room.setRevolution(revolution == 1)
	insertRoomInfo(room)
	res = *room.GetRoomID()
	return res
}

//将房间信息同步到负载均衡服务器
func insertRoomInfo(room *Room) {
	xclient := frame.NewRpcxClient(frame.GetArgs().RpcxClient)
	defer xclient.Close()
	args := &rpcs.Args{}
	args.V = map[string]interface{}{"args": []string{*room.GetRoomID(), addr}}
	reply := &rpcs.Reply{}
	err := xclient.Call(context.Background(), "InsertRoomInfo", args, reply)
	logger.CheckError(err, "insertRoomInfo")
}

/*
通过roomid获取房间信息
in:roomid
out:比赛类型|房间类型
*/
func GetRoomInfoByRoomID(args []string) *string {
	res := "0"
	roomid := args[1]
	room := RoomManage.GetRoom(roomid)
	if room != nil {
		if room.GetMatchID() == Match_GJYX {
			res = fmt.Sprintf("%d|%d", room.GetMatchID(), room.GetRoomType())
		} else if room.GetMatchID() == Match_HYTW {
			res = fmt.Sprintf("%d|%s", room.GetMatchID(), *room.GetRoomID())
		}
	}
	return &res
}

/*
获取正在比赛的房间ID
in:
out:-101没在房间中
	>0 房间ID
*/
func GetMatchingRoomID(args []string) *string {
	userid := args[0]
	user := UserManage.GetUser(&userid)
	if user == nil {
		return &Res_Unknown
	}
	room := user.getRoom()
	if room == nil {
		return &Res_Unknown
	}
	return room.GetRoomID()
}

/*
获取比赛规则
in:
out:0正常,1录像
*/
func GetGameRule(args []string) *string {
	userid := args[0]
	user := UserManage.GetUser(&userid)
	if user == nil {
		return &Res_Unknown
	}
	room := user.getRoom()
	if room == nil {
		return &Res_Unknown
	}
	gameRule := strconv.Itoa(room.getGameRule())
	return &gameRule
}

/*
获取上班玩家的3信息
in:
out:userid,cardid,第二张3出现的位置
*/
func GetGTW(args []string) *string {
	userid := args[0]
	user := UserManage.GetUser(&userid)
	//	if user == nil {
	//		user = UserManage.GetJudgmentUser()
	//	}
	if user == nil {
		return &Res_Unknown
	}
	room := user.getRoom()
	res := room.getGTW()
	return &res
}

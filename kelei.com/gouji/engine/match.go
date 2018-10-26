/*
比赛
*/

package engine

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/smallnest/rpcx/client"

	. "kelei.com/utils/common"
	"kelei.com/utils/frame"
	"kelei.com/utils/logger"
	"kelei.com/utils/rpcs"
)

var (
	addr    string
	lock    sync.Mutex
	xclient client.XClient
)

func engine_Init(addr_ string) {
	addr = addr_
	xclient = frame.NewRpcxClient(frame.GetArgs().RpcxClient)
	removeUserMatchInfo()
	removeRoomMatchInfo()
	go func() {
		for {
			syncGameInfo() //同步服务器人数信息
			time.Sleep(time.Second)
		}
	}()
	go func() {
		for {
			syncPCount()     //同步所有比赛类型的人数信息
			savePCountToDB() //保存所有比赛类型的人数信息
			time.Sleep(time.Second * 5)
		}
	}()
}

func Reset() {
	mapPCount = make(map[int]map[int]int)
	RoomManage.Reset()
	UserManage.Reset()
}

//获得game-redis
func getGameRds() redis.Conn {
	rds := frame.GetRedis("game")
	return rds
}

//获得nm-redis
func getNMRds() redis.Conn {
	rds := frame.GetRedis("member")
	return rds
}

//给玩家列表推送信息(比赛中的推送)
func pushMessageToUsers(funcName string, messages []string, userids []string) {
	msgCount := len(messages)
	//消息的数量是多个,但是数量与要推送的人数不符
	if msgCount > 1 && len(userids) != msgCount {
		logger.Errorf("pushMessageToUsers : %s", "消息的数量是多个,但是数量与要推送的人数不符")
		return
	}
	for i, userid := range userids {
		if userid != "" {
			u := UserManage.GetUser(&userid)
			if u == nil {
				u = UserManage.GetJudgmentUser()
			}
			if u == nil {
				continue
			}
			message := ""
			if msgCount <= 1 {
				message = messages[0]
			} else {
				message = messages[i]
			}
			u.push(funcName, &message)
		}
	}
}

//删除数据库中此服务器下的用户比赛信息
func removeUserMatchInfo() {
	args := &rpcs.Args{}
	args.V = map[string]interface{}{"args": []string{addr}}
	reply := &rpcs.Reply{}
	err := xclient.Call(context.Background(), "RemoveUserMatchInfo", args, reply)
	logger.CheckError(err, "removeUserMatchInfo")
}

//删除数据库中此服务器下的用户比赛信息
func removeRoomMatchInfo() {
	args := &rpcs.Args{}
	args.V = map[string]interface{}{"args": []string{addr}}
	reply := &rpcs.Reply{}
	err := xclient.Call(context.Background(), "RemoveRoomMatchInfo", args, reply)
	logger.CheckError(err, "removeRoomMatchInfo")
}

var mapPCount = make(map[int]map[int]int)

//向负载均衡服务器同步此游戏服务器的“所有类型的比赛人数”
func syncPCount() {
	lock.Lock()
	defer lock.Unlock()
	pCountInfo := *getMatchPCountInfo()
	//此服务器没有房间
	if pCountInfo == Res_NoBack {
		return
	}
	args := &rpcs.Args{}
	args.V = map[string]interface{}{"args": []string{addr, pCountInfo}}
	reply := &rpcs.Reply{}
	err := xclient.Call(context.Background(), "SyncPCount", args, reply)
	if err != nil {
		logger.Errorf("syncPCount : %s", err.Error())
		return
	}
	strPCount := reply.RS
	strPCountToMapPcount(strPCount)
}

//获取比赛人数的信息
func getMatchPCountInfo() *string {
	res := Res_Unknown
	m := make(map[int]map[int]int)
	rooms := RoomManage.GetRooms()
	for _, room := range rooms {
		matchID, roomType := room.GetMatchID(), room.GetRoomType()
		if m[matchID] == nil {
			m[matchID] = make(map[int]int)
		}
		m[matchID][roomType] += (room.GetPCount() + room.GetIdlePCount())
	}
	fillMap(m)
	buff := bytes.Buffer{}
	for matchID, mapRoom := range m {
		for roomType, pcount := range mapRoom {
			buff.WriteString(fmt.Sprintf("%d$%d$%d|", matchID, roomType, pcount))
		}
	}
	str := buff.String()
	if str != "" {
		res = str[:len(str)-1]
	}
	return &res
}

//填充人数信息
func fillMap(m map[int]map[int]int) {
	allMatchInfo := *getAllMatchInfo()
	arrMatchInfo := strings.Split(allMatchInfo, "|")
	for _, matchInfo := range arrMatchInfo {
		info := strings.Split(matchInfo, "$")
		matchID, _ := strconv.Atoi(info[0])
		room := Room{}
		allRoomData := *room.getAllRoomData()
		arrAllRoomData := strings.Split(allRoomData, "|")
		for _, roomData := range arrAllRoomData {
			arrRoomData_s := strings.Split(roomData, "$")
			arrRoomData := StrArrToIntArr(arrRoomData_s)
			roomType := arrRoomData[0]
			if m[matchID] == nil {
				m[matchID] = make(map[int]int)
			}
			if m[matchID][roomType] == 0 {
				m[matchID][roomType] = 0
			}
		}
	}
}

//将字符串人数转化成map人数
func strPCountToMapPcount(strPCount *string) {
	if *strPCount == Res_NoBack {
		return
	}
	infos := strings.Split(*strPCount, "|")
	for _, info_s := range infos {
		info := StrArrToIntArr(strings.Split(info_s, "$"))
		matchID := info[0]
		roomType := info[1]
		pcount := info[2]
		if mapPCount[matchID] == nil {
			mapPCount[matchID] = make(map[int]int)
		}
		mapPCount[matchID][roomType] = pcount
	}
}

//将人数存入数据库
func savePCountToDB() {
	//map[matchID]pcount
	m := make(map[int]int)
	for matchID, mapRoom := range mapPCount {
		for _, pcount := range mapRoom {
			m[matchID] += pcount
		}
	}
	allMatchInfo := *getAllMatchInfo()
	arrMatchInfo := strings.Split(allMatchInfo, "|")
	arr := []string{}
	for _, matchInfo := range arrMatchInfo {
		info := strings.Split(matchInfo, "$")
		matchID, _ := strconv.Atoi(info[0])
		pcount := m[matchID]
		arr = append(arr, strconv.Itoa(matchID), strconv.Itoa(pcount))
	}
	pCountInfo := strings.Join(arr, ",")
	_, err := db.Exec("call PCountUpdate(?)", pCountInfo)
	logger.CheckFatal(err, "savePCountToDB")
}

//向负载均衡服务器同步此游戏服务器的人数和房间数
func syncGameInfo() {
	defer func() {
		if p := recover(); p != nil {
			errInfo := fmt.Sprintf("连接负载均衡服务器失败 : { %v }", p)
			logger.Errorf(errInfo)
		}
	}()
	lock.Lock()
	defer lock.Unlock()
	mapRoomInfo := map[string]int{}
	for _, room := range RoomManage.GetRooms() {
		roomkey := fmt.Sprintf("%s:%d:%d", addr, room.GetMatchID(), room.GetRoomType())
		mapRoomInfo[roomkey] = mapRoomInfo[roomkey] + room.GetPCount() + room.GetIdlePCount()
	}

	args_ := []string{}
	for roomkey, usercount := range mapRoomInfo {
		args_ = append(args_, fmt.Sprintf("%s:%d", roomkey, usercount))
	}

	args := &rpcs.Args{}
	args.V = map[string]interface{}{"args": args_}
	reply := &rpcs.Reply{}
	err := xclient.Call(context.Background(), "SyncGameInfo", args, reply)
	logger.CheckError(err, "syncGameInfo")
}

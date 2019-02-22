/*
比赛
*/

package engine

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/garyburd/redigo/redis"

	. "kelei.com/utils/common"
	"kelei.com/utils/frame"
	"kelei.com/utils/logger"
)

const (
	MATCHTYPECOUNT    = 5
	GJYXROOMTYPECOUNT = 5
)

var (
	addr         string
	lock         sync.Mutex
	pcountT      int     //总人数(所有服务器)
	matchpcountT []int   //某类型赛事人数(所有服务器)
	roompcountT  [][]int //某类型房间人数(所有服务器)
)

func engine_Init(addr_ string) {
	addr = addr_
	go func() {
		for {
			savePCount()
			mergePCount()
			time.Sleep(time.Second * 5)
		}
	}()
}

func Reset() {
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

//保存本服务器人数
func savePCount() {
	defer func() {
		if p := recover(); p != nil {
			errInfo := fmt.Sprintf("savePCount : { %v }", p)
			logger.Errorf(errInfo)
		}
	}()
	//服务器总人数
	pcount := UserManage.GetUserCount()
	//某类型赛事人数
	matchpcount := make([]int, MATCHTYPECOUNT)
	//某类型房间人数
	roompcount := make([][]int, len(matchpcount))
	//初始化房间类型
	roompcount[Match_GJYX] = make([]int, GJYXROOMTYPECOUNT)
	for _, room := range RoomManage.GetRooms() {
		matchid := room.GetMatchID()
		roomtype := room.GetRoomType()
		for _, user := range room.getUsers() {
			if user != nil {
				matchpcount[matchid] += 1
				roompcount[matchid][roomtype] += 1
			}
		}
	}
	rds := getGameRds()
	defer rds.Close()
	roompcount_ := []string{}
	for _, v := range roompcount {
		roompcount_ = append(roompcount_, strings.Join(IntArrToStrArr(v), "|"))
	}
	matchpcountStr := strings.Join(IntArrToStrArr(matchpcount), ",")
	roompcountStr := strings.Join(roompcount_, ",")
	rds.Do("hmset", addr, "pcount", pcount, "matchpcount", matchpcountStr, "roompcount", roompcountStr)
}

//合并所有服务器人数
func mergePCount() {
	defer func() {
		if p := recover(); p != nil {
			errInfo := fmt.Sprintf("mergePCount : { %v }", p)
			logger.Errorf(errInfo)
		}
	}()
	rds := getGameRds()
	defer rds.Close()
	servers, err := redis.Strings(rds.Do("lrange", "servers", 0, -1))
	logger.CheckError(err)
	//总人数(所有服务器)
	pcountT := 0
	//某类型赛事人数(所有服务器)
	matchpcountT = make([]int, MATCHTYPECOUNT)
	//某类型房间人数(所有服务器)
	roompcountT = make([][]int, len(matchpcountT))
	//初始化房间类型(所有服务器)
	roompcountT[Match_GJYX] = make([]int, GJYXROOMTYPECOUNT)
	for _, server := range servers {
		data, err := redis.Strings(rds.Do("hmget", server, "pcount", "matchpcount", "roompcount"))
		logger.CheckError(err)
		pcountT += StrToInt(data[0])
		matchpcount := data[1]
		arr := strings.Split(matchpcount, ",")
		for i, v := range arr {
			matchpcountT[i] += StrToInt(v)
		}
		roompcount := data[2]
		arr = strings.Split(roompcount, ",")
		for i, v := range arr {
			if v != "" {
				arr2 := strings.Split(v, "|")
				for j, vv := range arr2 {
					roompcountT[i][j] += StrToInt(vv)
				}
			}
		}
	}
}

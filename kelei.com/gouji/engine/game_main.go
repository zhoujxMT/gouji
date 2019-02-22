/*
游戏服务器操作-主页
*/

package engine

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/garyburd/redigo/redis"

	. "kelei.com/utils/common"
	"kelei.com/utils/logger"
)

/*
获取主场景信息
in:
out:玩家名|等级|vip等级|元宝|英雄币,比赛类型$人数|
des:比赛类型:0够级英雄,1好友同玩,2开点发四,3大奖赛
*/
func GetMainInfo(args []string) *string {
	userid := args[0]
	//创建玩家
	user := UserManage.GetUser(&userid)
	//头部信息
	heroCoin := user.getHeroCoin()
	userInfo, err2 := redis.Values(user.GetUserInfo("username", "ingot", "level", "vip"))
	logger.CheckFatal(err2, "GetMainInfo:2")
	username, _ := redis.String(userInfo[0], nil)
	ingot, _ := redis.Int(userInfo[1], nil)
	level, _ := redis.Int(userInfo[2], nil)
	vip, _ := redis.Int(userInfo[3], nil)
	//比赛类型信息
	allMatchInfo := getAllMatchInfo()
	res := fmt.Sprintf("%s|%d|%d|%d|%d,%s", username, level, vip, ingot, heroCoin, *allMatchInfo)
	return &res
}

/*
获取所有比赛类型的信息
in:
out:比赛类型$人数|
des:比赛类型:0够级英雄,1好友同玩,2开点发四,3大奖赛
*/
func getAllMatchInfo() *string {
	info := Res_Unknown
	key := "allMatchInfo"
	gameRds := getGameRds()
	defer gameRds.Close()
	fromDB := func() {
		buff := bytes.Buffer{}
		matchID, status := 0, 0
		rows, err := db.Query("select matchID,status from matchinfo")
		logger.CheckFatal(err, "getAllMatchInfo")
		defer rows.Close()
		for rows.Next() {
			rows.Scan(&matchID, &status)
			//比赛类型已开启
			if status == 1 {
				buff.WriteString(fmt.Sprintf("%d$%d|", matchID, matchpcountT[matchID]))
			}
		}
		info = *RemoveLastChar(buff)
		gameRds.Do("set", key, info)
		gameRds.Do("expire", key, expireSecond)
	}
	var err error
	info, err = redis.String(gameRds.Do("get", key))
	if err != nil {
		fromDB()
	}
	return &info
}

/*
进入某类型的比赛
in:比赛类型
out:-2比赛类型未开放
	1进入成功
*/
func EnterMatch(args []string) *string {
	res := "1"
	matchID, err := strconv.Atoi(args[1])
	logger.CheckFatal(err, "EnterMatch:1")
	matchInfo, err2 := redis.Ints(GetMatchInfo(matchID, "status"))
	logger.CheckFatal(err2, "EnterMatch:2")
	status := matchInfo[0]
	if status == 0 {
		res = "-2"
		return &res
	}
	return &res
}

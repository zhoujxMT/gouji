/*
房间数据库操作
*/

package engine

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/garyburd/redigo/redis"

	"kelei.com/utils/logger"
)

/*
=========================================================================================================
表-BalanceData
=========================================================================================================
*/

func (r *Room) getBalanceDataKey(balanceType int) string {
	return fmt.Sprintf("roomT:%d:blceT:%d", r.GetRoomType(), balanceType)
}

//加载表balancedata数据
func (r *Room) loadBalanceData() {
	gameRds := getGameRds()
	defer gameRds.Close()
	exists, err := redis.Int(gameRds.Do("exists", r.getBalanceDataKey(Ranking_One)))
	logger.CheckFatal(err, "loadBalanceData")
	//不存在就从数据库中取
	if exists == 0 {
		roomType, balanceType, ingot, integral := 0, 0, 0, 0
		rows, err := db.Query("select RoomType,BalanceType,ingot,integral from BalanceData where roomtype=?", r.GetRoomType())
		logger.CheckFatal(err, "loadBalanceData2")
		defer rows.Close()
		buff := bytes.Buffer{}
		for rows.Next() {
			rows.Scan(&roomType, &balanceType, &ingot, &integral)
			buff.WriteString(fmt.Sprintf("%d,%d,%d,%d", roomType, balanceType, ingot, integral))
			buff.WriteString("|")
		}
		str := buff.String()
		str = str[:len(str)-1]
		_, err2 := gameRds.Do("evalsha", luaScript, 1, "balancedata", str)
		logger.CheckFatal(err2)
	}
}

//获取表数据
func (r *Room) GetBalanceData(balanceType int, args ...interface{}) (interface{}, error) {
	r.loadBalanceData()
	args = append(args[:0], append([]interface{}{r.getBalanceDataKey(balanceType)}, args[0:]...)...)
	gameRds := getGameRds()
	defer gameRds.Close()
	return gameRds.Do("hmget", args...)
}

/*
=========================================================================================================
表-RoomData
=========================================================================================================
*/

func (r *Room) getAllRoomData() *string {
	res := ""
	key := "allRoomData"
	gameRds := getGameRds()
	defer gameRds.Close()
	fromDB := func() {

		buff := bytes.Buffer{}
		roomtype, enteringot, level, multiple, matchingingot, expendingot := 0, 0, 0, 0, 0, 0
		rows, err := db.Query("select roomtype,enteringot,level,multiple,matchingingot,expendingot from roomdata")
		logger.CheckFatal(err, "getAllRoomData")
		defer rows.Close()
		for rows.Next() {
			rows.Scan(&roomtype, &enteringot, &level, &multiple, &matchingingot, &expendingot)
			buff.WriteString(fmt.Sprintf("%d$%d$%d$%d$%d$%d|", roomtype, enteringot, multiple, matchingingot, expendingot, level))
		}
		res = buff.String()
		res = res[:len(res)-1]
		gameRds.Do("set", key, res)
		gameRds.Do("expire", key, expireSecond)
	}
	var err error
	res, err = redis.String(gameRds.Do("get", key))
	if err != nil {
		fromDB()
	}
	return &res
}

func (r *Room) getRoomDataKey() string {
	return fmt.Sprintf("roomdata:roomT:%d", r.GetRoomType())
}

//加载表roomdata数据
func (r *Room) loadRoomData() {
	key := r.getRoomDataKey()
	gameRds := getGameRds()
	defer gameRds.Close()
	exists, err := redis.Int(gameRds.Do("exists", key))
	logger.CheckFatal(err, "loadRoomData")
	//不存在就从数据库中取
	if exists == 0 {
		playIngot, enterIngot, level, multiple, matchingIngot, winExp, loseExp, expendIngot, integral, charm, fleeIngot, fleeIntegral := 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0
		row := db.QueryRow("select PlayIngot,EnterIngot,level,multiple,MatchingIngot,WinExp,LoseExp,ExpendIngot,Integral,Charm,fleeIngot, fleeIntegral from RoomData where roomtype=?", r.GetRoomType())
		err = row.Scan(&playIngot, &enterIngot, &level, &multiple, &matchingIngot, &winExp, &loseExp, &expendIngot, &integral, &charm, &fleeIngot, &fleeIntegral)
		logger.CheckFatal(err, "loadRoomData2")
		_, err = gameRds.Do("hmset", key, "playIngot", playIngot, "enterIngot", enterIngot, "level", level, "multiple", multiple, "matchingIngot", matchingIngot, "winExp", winExp, "loseExp", loseExp, "expendIngot", expendIngot, "integral", integral, "charm", charm, "fleeIngot", fleeIngot, "fleeIntegral", fleeIntegral)
		logger.CheckFatal(err, "loadRoomData3")
		gameRds.Do("expire", key, expireSecond)
	}
}

//获取表数据
func (r *Room) GetRoomData(args ...interface{}) (interface{}, error) {
	r.loadRoomData()
	args = append(args[:0], append([]interface{}{r.getRoomDataKey()}, args[0:]...)...)
	gameRds := getGameRds()
	defer gameRds.Close()
	return gameRds.Do("hmget", args...)
}

/*
=========================================================================================================
表-MatchData
=========================================================================================================
*/

func (r *Room) getMatchDataKey() string {
	return fmt.Sprintf("matchdata:%d", r.GetMatchID())
}

//获得比赛类型的配置信息
func (r *Room) GetMatchData() *string {
	gameRds := getGameRds()
	defer gameRds.Close()
	key := r.getMatchDataKey()
	info, err := redis.String(gameRds.Do("get", key))
	if err != nil {

		err := db.QueryRow("select content from matchdata where id=?", r.GetMatchID()).Scan(&info)
		logger.CheckFatal(err, "GetMatchData")
		gameRds.Do("set", key, info)
		gameRds.Do("expire", key, expireSecond_Long)
	}
	return &info
}

/*
=========================================================================================================
表-GameRule
=========================================================================================================
*/

func (r *Room) getGameRuleKey() string {
	return fmt.Sprintf("gamerule")
}

//获取游戏规则
func (r *Room) GetGameRuleConfig() int {
	gameRds := getGameRds()
	defer gameRds.Close()
	key := r.getGameRuleKey()
	info, err := redis.String(gameRds.Do("get", key))
	if err != nil {

		err := db.QueryRow("select value from config where name=?", "gamerule").Scan(&info)
		logger.CheckFatal(err, "getGameRule")
		gameRds.Do("set", key, info)
		gameRds.Do("expire", key, expireSecond_Long)
	}
	gameRule, _ := strconv.Atoi(info)
	return gameRule
}

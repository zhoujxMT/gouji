/*
比赛信息
*/

package engine

import (
	"fmt"

	"github.com/garyburd/redigo/redis"

	"kelei.com/utils/frame"
	"kelei.com/utils/logger"
)

func getMatchInfoKey(matchID int) string {
	return fmt.Sprintf("matchinfo:%d", matchID)
}

//加载matchinfo信息到game-redis中
func loadMatchInfo(matchID int) {
	key := getMatchInfoKey(matchID)
	gameRds := getGameRds()
	defer gameRds.Close()
	exists, err := redis.Int(gameRds.Do("exists", key))
	logger.CheckFatal(err, "loadMatchInfo")
	if exists == 0 {
		db := frame.GetDB("game")
		status, count := 0, 0
		err := db.QueryRow("select status,pcount from matchinfo where matchid=?", matchID).Scan(&status, &count)
		logger.CheckFatal(err, "loadMatchInfo")
		gameRds.Do("hmset", key, "status", status, "pcount", count)
		gameRds.Do("expire", key, expireSecond)
	}
}

/*
获取比赛信息
*/
func GetMatchInfo(matchID int, args ...interface{}) (interface{}, error) {
	loadMatchInfo(matchID)
	args = append(args[:0], append([]interface{}{getMatchInfoKey(matchID)}, args[0:]...)...)
	gameRds := getGameRds()
	defer gameRds.Close()
	return gameRds.Do("hmget", args...)
}

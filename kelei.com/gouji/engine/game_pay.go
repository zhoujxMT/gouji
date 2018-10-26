/*
游戏服务器操作-充值
*/

package engine

import (
	"bytes"
	"fmt"

	"github.com/garyburd/redigo/redis"

	. "kelei.com/utils/common"
	"kelei.com/utils/logger"
)

//获取payinfo的key
func getPayInfoKey() string {
	return "payinfo"
}

/*
获取充值信息
in:
out:人民币$多送百分比$总英雄币$总元宝|
*/
func GetPayInfo(args []string) *string {
	gameRds := getGameRds()
	defer gameRds.Close()
	info, err := redis.String(gameRds.Do("get", getPayInfoKey()))
	if err != nil {

		rows, err := db.Query("select rmb,herocoin,ingot,extrapercent from paydata")
		logger.CheckFatal(err, "GetPayInfo")
		defer rows.Close()
		buff := bytes.Buffer{}
		rmb, heroCoin, ingot, extraPercent := 0, 0, 0, 0
		for rows.Next() {
			rows.Scan(&rmb, &heroCoin, &ingot, &extraPercent)
			percent := 1 + float32(extraPercent)/100
			heroCoin = int(float32(heroCoin) * percent)
			ingot = int(float32(ingot) * percent)
			buff.WriteString(fmt.Sprintf("%d$%d$%d$%d|", rmb, extraPercent, heroCoin, ingot))
		}
		info = *RemoveLastChar(buff)
		gameRds.Do("set", getPayInfoKey(), info)
		gameRds.Do("expire", getPayInfoKey(), expireSecond)
	}
	if info == "" {
		info = Res_Unknown
	}
	return &info
}

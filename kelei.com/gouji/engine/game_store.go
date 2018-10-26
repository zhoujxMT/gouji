/*
游戏服务器操作-商城
*/

package engine

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"

	. "kelei.com/utils/common"
	"kelei.com/utils/logger"
)

//获取storeinfo的key
func getStoreInfoKey(storeType string) string {
	return "storeinfo:" + storeType
}

/*
获取商城信息
in:商城类型(0英雄币商城 1元宝商城)
out:storeid$itemid$道具数量$价格$折扣(1-10)$限购(0不限购1个人限购2全服限购)$限购剩余数量$是否是新品$是否限时$限时剩余秒数|
*/
func GetStoreInfo(args []string) *string {
	userid := args[0]
	storeType := args[1]
	key := getStoreInfoKey(storeType)
	gameRds := getGameRds()
	defer gameRds.Close()
	info, err := redis.String(gameRds.Do("get", key))
	if err != nil {

		rows, err := db.Query("select storeID,itemID,itemCount,price,timeLimit,upTime,downTime,discount,quotaType,quotaCount from storeinfo where storetype=?  and uptime<NOW() and downtime>NOW()", storeType)
		logger.CheckFatal(err, "GetStoreInfo")
		defer rows.Close()
		buff := bytes.Buffer{}
		upTime_s, downTime_s := "", ""
		timeLimit_u := []uint8{}
		storeID, itemID, itemCount, price, discount, quotaType, quotaCount := 0, 0, 0, 0, 0, 0, 0
		for rows.Next() {
			err := rows.Scan(&storeID, &itemID, &itemCount, &price, &timeLimit_u, &upTime_s, &downTime_s, &discount, &quotaType, &quotaCount)
			logger.CheckFatal(err)
			//是否是新品
			isNew := 0
			upTime, _ := time.ParseInLocation("2006-01-02 15:04:05", upTime_s, time.Local)
			if time.Now().Sub(upTime) < time.Hour*72 {
				isNew = 1
			}
			//是否限时
			isTimeLimit := int(timeLimit_u[0])
			if isTimeLimit == 0 {
				downTime_s = ""
			}
			//折扣
			price = int(float32(price) * float32(discount) / 10)
			buff.WriteString(fmt.Sprintf("%d$%d$%d$%d$%d$%d$%d$%d$%s$%d|", storeID, itemID, itemCount, price, discount, quotaType, isNew, isTimeLimit, downTime_s, quotaCount))
		}
		info = *RemoveLastChar(buff)
		gameRds.Do("set", key, info)
		gameRds.Do("expire", key, expireSecond)
	}
	if info == "" {
		info = Res_Unknown
	} else {
		buff := bytes.Buffer{}
		arrStoreInfo := strings.Split(info, "|")
		var storeID, itemID, itemCount, price, discount, quotaType, isNew, isTimeLimit, downTime_s, quotaCount_s string
		for _, storeInfo := range arrStoreInfo {
			arrStoreInfo := strings.Split(storeInfo, "$")
			quotaSurplusCount := 0
			storeID, itemID, itemCount, price, discount, quotaType, isNew, isTimeLimit, downTime_s, quotaCount_s = arrStoreInfo[0], arrStoreInfo[1], arrStoreInfo[2], arrStoreInfo[3], arrStoreInfo[4], arrStoreInfo[5], arrStoreInfo[6], arrStoreInfo[7], arrStoreInfo[8], arrStoreInfo[9]
			//限购
			if quotaType != "0" {
				if quotaType == "2" {
					userid = "0"
				}
				quotaCount, _ := strconv.Atoi(quotaCount_s)
				quotaSurplusCount = quotaCount - getUserBuyCount(userid, storeID)
			}
			//限时倒计时
			timeLimitSurplusSeconds := 0
			if isTimeLimit == "1" {
				downTime, _ := time.ParseInLocation("2006-01-02 15:04:05", downTime_s, time.Local)
				timeLimitSurplusSeconds = int(downTime.Sub(time.Now()).Seconds())
				if timeLimitSurplusSeconds <= 0 {
					continue
				}
			}
			buff.WriteString(fmt.Sprintf("%s$%s$%s$%s$%s$%s$%d$%s$%s$%d|", storeID, itemID, itemCount, price, discount, quotaType, quotaSurplusCount, isNew, isTimeLimit, timeLimitSurplusSeconds))
		}
		info = *RemoveLastChar(buff)
	}
	return &info
}

//获取storeinfo的key
func getLogStoreKey(userid string, storeid string) string {
	return "logstore:" + userid + ":" + storeid
}

/*
获取玩家购买的数量
userid:0是全服限购 其它个人限购
*/
func getUserBuyCount(userid string, storeid string) int {
	key := getLogStoreKey(userid, storeid)
	gameRds := getGameRds()
	defer gameRds.Close()
	count, err := redis.Int(gameRds.Do("get", key))
	if err != nil {

		err := db.QueryRow("select count from log_store where userid=? and storeid=?", userid, storeid).Scan(&count)
		logger.CheckFatal(err, "getUserBuyCount")
		gameRds.Do("set", key, count)
		gameRds.Do("expire", key, expireSecond)
	}
	return count
}

/*
商城购买
in:商城类型(0英雄币商城 1元宝商城),storeid
out:状态(0购买成功 -1币不足 -2已下架 -3超过限购次数)|剩余币数量|限购剩余数量
*/
func Buy(args []string) *string {
	res := "-2"
	userid := args[0]
	storetype := args[1]
	storeid := args[2]

	gameRds := getGameRds()
	defer gameRds.Close()
	user := User{}
	user.setUserID(&userid)
	coin := 0
	info := *GetStoreInfo([]string{userid, storetype})
	arrStoreInfo := strings.Split(info, "|")
	for _, storeInfo := range arrStoreInfo {
		arrStoreInfo := strings.Split(storeInfo, "$")
		storeID, itemID, itemCount, price_s, discount_s, quotaType, quotaSurplusCount_s := arrStoreInfo[0], arrStoreInfo[1], arrStoreInfo[2], arrStoreInfo[3], arrStoreInfo[4], arrStoreInfo[5], arrStoreInfo[6]
		itemID, itemCount, discount_s = itemID, itemCount, discount_s
		if storeid == storeID {
			price, _ := strconv.Atoi(price_s)
			//			discount, _ := strconv.ParseFloat(discount_s, 32)
			//			discount = discount / 10
			//			price = int(float64(price) * discount)
			//获取币
			if storetype == "0" {
				coin = user.getHeroCoin()
			} else {
				coin = user.getIngot()
			}
			//币不足
			if price > coin {
				res = fmt.Sprintf("%d|%d|%d", -1, coin, 0)
				break
			}
			//限购
			if quotaType != "0" {
				quotaSurplusCount, _ := strconv.Atoi(quotaSurplusCount_s)
				if quotaSurplusCount <= 0 {
					res = fmt.Sprintf("%d|%d|%d", -3, coin, 0)
					break
				}
			}
			//扣除币
			if storetype == "0" {
				coin = user.updateHeroCoin(-price, 1)
			} else {
				coin = user.updateIngot(-price, 2)
			}
			//商城购买
			quotaSurplusCount := 0
			storeID_int, _ := strconv.Atoi(storeID)
			db.QueryRow("call StoreBuy(?,?)", userid, storeID_int).Scan(&quotaSurplusCount)
			//限购
			if quotaType == "0" {
				quotaSurplusCount = 0
			}
			res = fmt.Sprintf("%d|%d|%d", 0, coin, quotaSurplusCount)
			//设置键过期
			gameRds.Do("expire", getLogStoreKey(userid, storeid), 0)
			break
		}
	}
	return &res
}

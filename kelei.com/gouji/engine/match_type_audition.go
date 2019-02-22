package engine

import (
	"bytes"
	"fmt"
	"time"

	"github.com/garyburd/redigo/redis"

	"kelei.com/utils/common"
	"kelei.com/utils/json"
	"kelei.com/utils/logger"
)

var WeekDayMap = map[string]int{
	"Monday":    1,
	"Tuesday":   2,
	"Wednesday": 3,
	"Thursday":  4,
	"Friday":    5,
	"Saturday":  6,
	"Sunday":    0,
}

/*
获取海选赛排名
in:
out:赛季编号|轮次|星期几|当前时间#userid|玩家名|晋级券数量|状态|头像地址$#我的排名|我的晋级券数量|本轮剩余次数|我的状态
des:状态(0未报名1已报名2已晋级)
*/
func GetAuditionInfo(args []string) *string {
	userid := args[0]
	gameRds := getGameRds()
	defer gameRds.Close()
	key := "AuditionInfo"
	info, err := redis.String(gameRds.Do("get", key))
	if err != nil {
		timeInfo := func() string {
			room := Room{}
			room.setMatchID(Match_HXS)
			content := room.GetMatchData()
			hxs := getHXSContent(content)
			return fmt.Sprintf("%d|%d|%d|%s", hxs.SeasonIndex, hxs.Round, WeekDayMap[time.Now().Weekday().String()], time.Now().Format("2006-01-02 15:04:05"))
		}
		rankingInfo := func() string {
			rows, err := db.Query("call GetAuditionRanking()")
			logger.CheckFatal(err, "GetAuditionRanking")
			defer rows.Close()
			buff := bytes.Buffer{}
			userid, tickerCount, status := 0, 0, 0
			userName, headUrl := "", ""
			for rows.Next() {
				status = 0
				rows.Scan(&userid, &userName, &tickerCount, &headUrl, &status)
				buff.WriteString(fmt.Sprintf("%d|%s|%d|%d|%s$", userid, userName, tickerCount, status, headUrl))
			}
			return *common.RowsBufferToString(buff)
		}
		info = fmt.Sprintf("%s#%s", timeInfo(), rankingInfo())
		gameRds.Do("set", key, info)
		gameRds.Do("expire", key, expireSecond)
	}
	info = fmt.Sprintf("%s#%s", info, *GetMyRankingInfo(userid))
	return &info
}

func GetMyRankingInfo(userid string) *string {
	gameRds := getGameRds()
	defer gameRds.Close()
	key := "AuditionRanking:" + userid
	info, err := redis.String(gameRds.Do("get", key))
	if err != nil {
		myRankingInfo := func() string {
			ranking, tickerCount, surplusCount, auditionMatchTimeStatus, status := 0, 0, 0, 0, 0
			db.QueryRow("call GetAuditionMyRanking(?)", userid).Scan(&ranking, &tickerCount, &surplusCount, &auditionMatchTimeStatus, &status)
			return fmt.Sprintf("%d|%d|%d|%d", ranking, tickerCount, surplusCount, status)
		}
		info = myRankingInfo()
		gameRds.Do("set", key, info)
		gameRds.Do("expire", key, expireSecond)
	}
	return &info
}

//好友同玩
type HXS struct {
	SeasonIndex int
	Round       int
}

func getHXSContent(content *string) *HXS {
	hxs := HXS{}
	js := json.NewJsonStruct()
	js.LoadByString(*content, &hxs)
	return &hxs
}

/*
获取实名认证信息
in:
out:手机号|姓名|是否参加
des:是否参加(0不参加 >0参加)
*/
func GetCertification(args []string) *string {
	userid := args[0]
	gameRds := getGameRds()
	defer gameRds.Close()
	key := "Certification:" + userid
	info, err := redis.String(gameRds.Do("get", key))
	if err != nil {
		phonenumber, username := "", ""
		status := 0
		err := db.QueryRow("select phonenumber,username,status from Certification where userid=?", userid).Scan(&phonenumber, &username, &status)
		if err == nil {
			info = fmt.Sprintf("%s|%s|%d", phonenumber, username, status)
		} else {
			info = "||1"
		}
		gameRds.Do("set", key, info)
		gameRds.Do("expire", key, expireSecond_Long_Long)
	}
	return &info
}

/*
保存实名认证信息
in:手机号,姓名,是否参加
out:-1此时间段不可修改实名认证信息
	1成功
*/
func SaveCertification(args []string) *string {
	res := common.Res_Succeed
	userid := args[0]
	user := UserManage.GetUser(&userid)
	phoneNumber := args[1]
	userName := args[2]
	status := args[3]
	err := db.QueryRow("call SaveCertification(?,?,?,?)", *user.getUserID(), phoneNumber, userName, status).Scan(&res)
	logger.CheckError(err, "SaveCertification")
	gameRds := getGameRds()
	defer gameRds.Close()
	key := "Certification:" + userid
	gameRds.Do("expire", key, 0)
	return &res
}

/*
获取海选赛奖励
in:
out:itemid|count$itemid|count#
*/
func GetAuditionAward(args []string) *string {
	gameRds := getGameRds()
	defer gameRds.Close()
	key := "AuditionAward"
	info, err := redis.String(gameRds.Do("get", key))
	if err != nil {
		rows, err := db.Query("select content from AuditionAward")
		logger.CheckError(err)
		defer rows.Close()
		buff := bytes.Buffer{}
		content := ""
		for rows.Next() {
			rows.Scan(&content)
			buff.WriteString(fmt.Sprintf("%s#", content))
		}
		info = *common.RowsBufferToString(buff)
		gameRds.Do("set", key, info)
		gameRds.Do("expire", key, expireSecond_Long)
	}
	return &info
}

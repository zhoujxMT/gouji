package modes

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/gin-gonic/gin"

	. "kelei.com/utils/common"
	"kelei.com/utils/frame"
	"kelei.com/utils/logger"
)

type Enterprise struct {
}

type Element struct {
	Name string  `json:"name"`
	Y    float64 `json:"y"`
}

type Result_ShowEnterprise struct {
	StockName   string `json:"stockName"`
	StockCode   string `json:"stockCode"`
	TheirTrade  string `json:"theirTrade"`
	TotalEquity string
	Shares      string `json:"shares"`
	MarketValue string `json:"marketValue"`
	NowTime     string `json:"nowTime"`
	Weights     string
	Constitute  []Element `json:"constitute"`
}

func (e *Enterprise) ShowEnterprise(c *gin.Context, args []string) {
	id := args[0]
	enterprise := *e.getEnterprise(id)
	if enterprise == "" {
		logger.Errorf("企业不存在")
		return
	}
	info := strings.Split(enterprise, "|")
	stockName, stockCode, theirTrade, totalEquity, shares, marketValue, nowTime, weights := info[0], info[1], info[2], info[3], info[4], info[5], info[6], info[11]
	constitute_ := *e.getConstitute(id)
	if constitute_ == "" {
		logger.Errorf("企业不存在")
		return
	}
	arrElementInfo := strings.Split(constitute_, "|")
	constitute := []Element{}
	for _, elementInfo := range arrElementInfo {
		arrElement := strings.Split(elementInfo, "$")
		name, percent_s := arrElement[0], arrElement[1]
		percent, _ := strconv.ParseFloat(percent_s, 10)
		constitute = append(constitute, Element{name, percent})
	}
	res := Result_ShowEnterprise{stockName, stockCode, theirTrade, totalEquity, shares, marketValue, nowTime, weights, constitute}
	c.JSON(http.StatusOK, res)
}

//获取企业结构
func (e *Enterprise) getConstitute(id string) *string {
	rds := frame.GetRedis()
	defer rds.Close()
	key := fmt.Sprintf("constitute:%s", id)
	info, err := redis.String(rds.Do("get", key))
	if err != nil {
		db := frame.GetDB()
		rows, err := db.Query("select name,percent from constitute where enterpriseid=?;", id)
		logger.CheckFatal(err, "getConstitute")
		defer rows.Close()
		buff := bytes.Buffer{}
		var name string
		var percent string
		for rows.Next() {
			rows.Scan(&name, &percent)
			buff.WriteString(fmt.Sprintf("%s$%s|", name, percent))
		}
		info = *RemoveLastChar(buff)
		rds.Do("set", key, info)
		rds.Do("expire", key, ExpTime_Hour)
	}
	return &info
}

func (e *Enterprise) getEnterprise(id string) *string {
	rds := frame.GetRedis()
	defer rds.Close()
	key := fmt.Sprintf("enterprise:%s", id)
	info, err := redis.String(rds.Do("get", key))
	if err != nil {
		db := frame.GetDB()
		var stockName, stockCode, theirTrade, totalEquity, shares, marketValue, nowTime, investCount, growthSpacePremium, growthEfficiencyPremium, currentPriceToBook, weights string
		row := db.QueryRow("select stockName,stockCode,theirTrade,totalEquity,shares,marketValue,nowTime,investCount,growthSpacePremium,growthEfficiencyPremium,currentPriceToBook,weights from enterprise where id=?;", id)
		err := row.Scan(&stockName, &stockCode, &theirTrade, &totalEquity, &shares, &marketValue, &nowTime, &investCount, &growthSpacePremium, &growthEfficiencyPremium, &currentPriceToBook, &weights)
		logger.CheckFatal(err, "getEnterprise")
		t, _ := time.Parse("2006-01-02 15:04:05", nowTime)
		nowTime = t.Format("2006-01-02")
		buff := bytes.Buffer{}
		buff.WriteString(fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s", stockName, stockCode, theirTrade, totalEquity, shares, marketValue, nowTime, investCount, growthSpacePremium, growthEfficiencyPremium, currentPriceToBook, weights))
		info = buff.String()
		rds.Do("set", key, info)
		rds.Do("expire", key, ExpTime_Hour)
	}
	return &info
}

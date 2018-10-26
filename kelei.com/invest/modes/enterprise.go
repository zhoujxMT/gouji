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
	StockName           string `json:"stockName"`
	StockCode           string `json:"stockCode"`
	TheirTrade          string `json:"theirTrade"`
	TotalEquity         string
	Shares              string `json:"shares"`
	MarketValue         string `json:"marketValue"`
	NowTime             string `json:"nowTime"`
	Weights             string
	GrowthSpaceAnalysis string
	GrowthSpaceScore    string
	Constitute          []Element `json:"constitute"`
}

func (e *Enterprise) ShowEnterprise(c *gin.Context, args []string) {
	id := args[0]
	enterprise := *e.getEnterprise(id)
	if enterprise == "" {
		logger.Errorf("企业不存在")
		return
	}
	info := strings.Split(enterprise, "|")
	stockName, stockCode, theirTrade, totalEquity, shares, marketValue, nowTime, weights, growthSpaceAnalysis, growthSpaceScore := info[0], info[1], info[2], info[3], info[4], info[5], info[6], info[11], info[12], info[13]
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
	res := Result_ShowEnterprise{stockName, stockCode, theirTrade, totalEquity, shares, marketValue, nowTime, weights, growthSpaceAnalysis, growthSpaceScore, constitute}
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
		rds.Do("expire", key, ExpTime_Short)
	}
	return &info
}

func (e *Enterprise) getKey(id string) string {
	key := fmt.Sprintf("enterprise:%s", id)
	return key
}

func (e *Enterprise) getEnterprise(id string) *string {
	rds := frame.GetRedis()
	defer rds.Close()
	key := e.getKey(id)
	info, err := redis.String(rds.Do("get", key))
	if err != nil {
		db := frame.GetDB()
		var stockName, stockCode, theirTrade, totalEquity, shares, marketValue, nowTime, growthSpacePremium, growthEfficiencyPremium, currentPriceToBook, weights, growthSpaceAnalysis, growthSpaceScore, spaceEfficiencyDes string
		row := db.QueryRow("select stockName,stockCode,theirTrade,totalEquity,shares,marketValue,nowTime,growthSpacePremium,growthEfficiencyPremium,currentPriceToBook,weights,growthSpaceAnalysis,growthSpaceScore,spaceEfficiencyDes from enterprise where id=?;", id)
		err := row.Scan(&stockName, &stockCode, &theirTrade, &totalEquity, &shares, &marketValue, &nowTime, &growthSpacePremium, &growthEfficiencyPremium, &currentPriceToBook, &weights, &growthSpaceAnalysis, &growthSpaceScore, &spaceEfficiencyDes)
		logger.CheckFatal(err, "getEnterprise")
		t, _ := time.Parse("2006-01-02 15:04:05", nowTime)
		nowTime = t.Format("2006-01-02")
		investCount := e.getInvestCount(id)
		buff := bytes.Buffer{}
		buff.WriteString(fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s|%d|%s|%s|%s|%s|%s|%s|%s", stockName, stockCode, theirTrade, totalEquity, shares, marketValue, nowTime, investCount, growthSpacePremium, growthEfficiencyPremium, currentPriceToBook, weights, growthSpaceAnalysis, growthSpaceScore, spaceEfficiencyDes))
		info = buff.String()
		rds.Do("set", key, info)
		rds.Do("expire", key, ExpTime_Short)
	}
	return &info
}

//获取投资次数
func (e *Enterprise) getInvestCount(id string) int {
	rds := frame.GetRedis()
	defer rds.Close()
	key := "investCount"
	investCount, err := redis.Int(rds.Do("get", key))
	if err != nil {
		db := frame.GetDB()
		rows, err := db.Query("select * from (select year,InvestReceCash,LessShareholderInvestReceCash from enterprisedata where enterpriseid=? order by year desc limit 5) aa order by year", id)
		logger.CheckFatal(err, "getEnterpriseData")
		defer rows.Close()
		var year, investReceCash, lessShareholderInvestReceCash float64
		for rows.Next() {
			rows.Scan(&year, &investReceCash, &lessShareholderInvestReceCash)
			if investReceCash-lessShareholderInvestReceCash > 0 {
				investCount++
			}
		}
	}
	return investCount
}

type Result_UpdateGrowthSpace struct {
	Status string
}

//修改增长空间
func (e *Enterprise) UpdateGrowthSpace(c *gin.Context, args []string) {
	enterpriseID, growthSpaceAnalysis, growthSpaceScore := args[0], args[1], args[2]
	db := frame.GetDB()
	_, err := db.Exec("update enterprise set GrowthSpaceAnalysis=?,GrowthSpaceScore=? where id=?", growthSpaceAnalysis, growthSpaceScore, enterpriseID)
	logger.CheckError(err)
	key := e.getKey(enterpriseID)
	rds := frame.GetRedis()
	defer rds.Close()
	rds.Do("expire", key, 0)
	res := Result_UpdateGrowthSpace{Status: SC_OK}
	c.JSON(http.StatusOK, res)
}

//修改估值
func (e *Enterprise) UpdateAssess(c *gin.Context, args []string) {
	enterpriseID, spaceEfficiency, spacePremium, efficiencyPremium := args[0], args[1], args[2], args[3]
	fmt.Println(enterpriseID, spaceEfficiency, spacePremium, efficiencyPremium)
	db := frame.GetDB()
	_, err := db.Exec("update enterprise set SpaceEfficiencyDes=?,GrowthSpacePremium=?,GrowthEfficiencyPremium=? where id=?", spaceEfficiency, spacePremium, efficiencyPremium, enterpriseID)
	logger.CheckError(err)
	key := e.getKey(enterpriseID)
	rds := frame.GetRedis()
	defer rds.Close()
	rds.Do("expire", key, 0)
	res := Result_UpdateGrowthSpace{Status: SC_OK}
	c.JSON(http.StatusOK, res)
}

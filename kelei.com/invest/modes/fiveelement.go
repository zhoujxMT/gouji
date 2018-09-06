/*
盈利能力
*/

package modes

import (
	"bytes"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/garyburd/redigo/redis"
	"github.com/gin-gonic/gin"

	. "kelei.com/utils/common"
	"kelei.com/utils/frame"
	"kelei.com/utils/logger"
)

type Profitability struct {
}

type JsonProfitability struct {
	InvestCount             int     //近5年投资次数
	GrowthSpacePremium      float64 //增长空间溢价
	GrowthEfficiencyPremium float64 //增长效率溢价
	CurrentPriceToBook      float64 //当前市净率
	EnterpriseData          string  //企业全部数据
}

func (p *Profitability) ShowProfitability(c *gin.Context, args []string) {
	id := args[0]
	investCount, growthSpacePremium, growthEfficiencyPremium, currentPriceToBook := p.getEnterpriseInfo(id)
	//获取所有数据
	enterpriseData := p.getEnterpriseData(id)
	//生成json
	jsonPro := JsonProfitability{investCount, growthSpacePremium, growthEfficiencyPremium, currentPriceToBook, *enterpriseData}
	//若返回json数据，可以直接使用gin封装好的JSON方法
	c.JSON(http.StatusOK, jsonPro)
}

//获取企业信息
func (p *Profitability) getEnterpriseInfo(id string) (int, float64, float64, float64) {
	//获取投资次数
	enterprise := Enterprise{}
	info := *enterprise.getEnterprise(id)
	if info == "" {
		logger.Errorf("企业不存在")
		return 0, 0, 0, 0
	}
	infoArr := strings.Split(info, "|")
	investCount, _ := strconv.Atoi(infoArr[7])
	growthSpacePremium, _ := strconv.ParseFloat(infoArr[8], 10)
	growthEfficiencyPremium, _ := strconv.ParseFloat(infoArr[9], 10)
	currentPriceToBook, _ := strconv.ParseFloat(infoArr[10], 10)
	return investCount, growthSpacePremium, growthEfficiencyPremium, currentPriceToBook
}

func (p *Profitability) getKey(id string) string {
	return fmt.Sprintf("enterpriseData:%s", id)
}

//获得净利润数据
func (p *Profitability) getEnterpriseData(id string) *string {
	logger.Infof("[加载数据]")
	rds := frame.GetRedis()
	defer rds.Close()
	key := p.getKey(id)
	info, err := redis.String(rds.Do("get", key))
	if err != nil {
		db := frame.GetDB()
		rows, err := db.Query("select year,gmnetmargin,gmNetMarginGrowthRate,roe,totalAssetGrowthRate,netAssetGrowthRate,openIncomeGrowthRate,netMargin,netMarginGrowthRate,manageCashNetAmount,investCashNetAmount,InvestReceCash,LessShareholderInvestReceCash,NetAsset,LiabilityWithInterestRate from enterprisedata where enterpriseid=?", id)
		logger.CheckFatal(err, "getEnterpriseData")
		defer rows.Close()
		var year int
		var gmNetMargin, gmNetMarginGrowthRate, roe, totalAssetGrowthRate, netAssetGrowthRate, openIncomeGrowthRate, netMargin, netMarginGrowthRate, manageCashNetAmount, investCashNetAmount, investReceCash, lessShareholderInvestReceCash, netAsset, liabilityWithInterestRate float64
		gmNetMarginGrowthRates, roes, investCashNetAmounts, netMargins := []float64{}, []float64{}, []float64{}, []float64{}
		strRows := []string{}
		for rows.Next() {
			rows.Scan(&year, &gmNetMargin, &gmNetMarginGrowthRate, &roe, &totalAssetGrowthRate, &netAssetGrowthRate, &openIncomeGrowthRate, &netMargin, &netMarginGrowthRate, &manageCashNetAmount, &investCashNetAmount, &investReceCash, &lessShareholderInvestReceCash, &netAsset, &liabilityWithInterestRate)
			gmNetMarginGrowthRates = append(gmNetMarginGrowthRates, gmNetMarginGrowthRate)
			roes = append(roes, roe)
			investCashNetAmounts = append(investCashNetAmounts, investCashNetAmount)
			netMargins = append(netMargins, netMargin)
			//筹资额度
			financingAmount := investReceCash - lessShareholderInvestReceCash
			strRows = append(strRows, fmt.Sprintf("%d|%.2f|%.2f|%.2f|%.2f|%.2f|%.2f|%.2f|%.2f|%.2f|%.2f|%.2f|%.2f|%.2f", year, gmNetMargin, gmNetMarginGrowthRate, roe, totalAssetGrowthRate, netAssetGrowthRate, openIncomeGrowthRate, netMargin, netMarginGrowthRate, manageCashNetAmount, investCashNetAmount, financingAmount, netAsset, liabilityWithInterestRate))
		}
		//归母净利润波动率
		gmNetMarginVolatilitys := p.getVolatility(gmNetMarginGrowthRates)
		//净资产收益率(roe)波动率
		roeVolatilitys := p.getVolatility(roes)
		//投资率
		investGates := p.getInvestGates(investCashNetAmounts, netMargins)
		//拼接进去
		buff := bytes.Buffer{}
		var gmNetMarginVolatility, roeVolatility, investGate float64
		for rowid, strRow := range strRows {
			gmNetMarginVolatility, roeVolatility, investGate = gmNetMarginVolatilitys[rowid], roeVolatilitys[rowid], investGates[rowid]
			buff.WriteString(fmt.Sprintf("%s|%.2f|%.2f|%.2f$", strRow, gmNetMarginVolatility, roeVolatility, investGate))
		}
		info = *RemoveLastChar(buff)
		rds.Do("set", key, info)
		rds.Do("expire", key, ExpTime_Hour)
	}
	return &info
}

//获取投资率
func (p *Profitability) getInvestGates(investCashNetAmounts, netMargins []float64) []float64 {
	investGates := []float64{}
	for k, investCashNetAmount := range investCashNetAmounts {
		netMargin := netMargins[k]
		investGate_s := fmt.Sprintf("%.2f", investCashNetAmount/netMargin)
		investGate, _ := strconv.ParseFloat(investGate_s, 10)
		investGates = append(investGates, investGate)
	}
	return investGates
}

//获取波动率
func (p *Profitability) getVolatility(growthRates []float64) []float64 {
	volatilitys := []float64{}
	for i, growthRate := range growthRates {
		var v float64 = 0
		if i > 0 {
			v = growthRate - growthRates[i-1]
			v = Round2(v, 2)
			v = math.Abs(v)
		}
		volatilitys = append(volatilitys, v)
	}
	return volatilitys
}

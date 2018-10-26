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
	GrowthSpaceAnalysis     string  //增长空间分析
	GrowthSpaceScore        float64 //增长空间评分
	SpaceEfficiencyDes      string  //空间效率描述
}

func (p *Profitability) ShowProfitability(c *gin.Context, args []string) {
	id := args[0]
	investCount, growthSpacePremium, growthEfficiencyPremium, currentPriceToBook, growthSpaceAnalysis, growthSpaceScore, SpaceEfficiencyDes := p.getEnterpriseInfo(id)
	//获取所有数据
	enterpriseData := p.getEnterpriseData(id)
	//生成json
	jsonPro := JsonProfitability{investCount, growthSpacePremium, growthEfficiencyPremium, currentPriceToBook, *enterpriseData, growthSpaceAnalysis, growthSpaceScore, SpaceEfficiencyDes}
	//若返回json数据，可以直接使用gin封装好的JSON方法
	c.JSON(http.StatusOK, jsonPro)
}

//获取企业信息
func (p *Profitability) getEnterpriseInfo(id string) (int, float64, float64, float64, string, float64, string) {
	//获取投资次数
	enterprise := Enterprise{}
	info := *enterprise.getEnterprise(id)
	if info == "" {
		logger.Errorf("企业不存在")
		return 0, 0, 0, 0, "", 0, ""
	}
	infoArr := strings.Split(info, "|")
	investCount, _ := strconv.Atoi(infoArr[7])
	growthSpacePremium, _ := strconv.ParseFloat(infoArr[8], 10)
	growthEfficiencyPremium, _ := strconv.ParseFloat(infoArr[9], 10)
	currentPriceToBook, _ := strconv.ParseFloat(infoArr[10], 10)
	growthSpaceAnalysis := infoArr[12]
	growthSpaceScore, _ := strconv.ParseFloat(infoArr[13], 10)
	spaceEfficiencyDes := infoArr[14]
	return investCount, growthSpacePremium, growthEfficiencyPremium, currentPriceToBook, growthSpaceAnalysis, growthSpaceScore, spaceEfficiencyDes
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
		rows, err := db.Query("select * from (select year,gmnetmargin,roe,totalAsset,gmNetAsset,openIncome,netMargin,netMarginGrowthRate,manageCashNetAmount,investCashNetAmount,InvestReceCash,lessShareholderInvestReceCash,NetAsset,ShortLoan,NoCLMWOY,LongLoan,LongClaim from enterprisedata where enterpriseid=? order by year desc limit 6) aa order by year", id)
		logger.CheckFatal(err, "getEnterpriseData")
		defer rows.Close()
		var year int
		var gmNetMargin, roe, totalAsset, gmNetAsset, openIncome, netMargin, netMarginGrowthRate, manageCashNetAmount, investCashNetAmount, investReceCash, lessShareholderInvestReceCash, netAsset, shortLoan, noCLMWOY, longLoan, longClaim float64
		gmNetMargins, roes, investCashNetAmounts, netMargins, netMarginGrowthRates, totalAssets, gmNetAssets, openIncomes := []float64{}, []float64{}, []float64{}, []float64{}, []float64{}, []float64{}, []float64{}, []float64{}
		strRows := []string{}
		rowid := 0
		for rows.Next() {
			rows.Scan(&year, &gmNetMargin, &roe, &totalAsset, &gmNetAsset, &openIncome, &netMargin, &netMarginGrowthRate, &manageCashNetAmount, &investCashNetAmount, &investReceCash, &lessShareholderInvestReceCash, &netAsset, &shortLoan, &noCLMWOY, &longLoan, &longClaim)
			gmNetMargins = append(gmNetMargins, gmNetMargin)
			roes = append(roes, roe)
			investCashNetAmounts = append(investCashNetAmounts, investCashNetAmount)
			netMargins = append(netMargins, netMargin)
			netMarginGrowthRates = append(netMarginGrowthRates, netMarginGrowthRate)
			totalAssets = append(totalAssets, totalAsset)
			gmNetAssets = append(gmNetAssets, gmNetAsset)
			openIncomes = append(openIncomes, openIncome)
			//筹资额度
			financingAmount := investReceCash - lessShareholderInvestReceCash
			//有息负债率
			liabilityWithInterestRate := (shortLoan + noCLMWOY + longLoan + longClaim) / netAsset * 100
			//			if rowid > 0 {
			strRows = append(strRows, fmt.Sprintf("%d|%.2f|%.2f|%.2f|%.2f|%.2f|%.2f|%.2f|%.2f", year, gmNetMargin, roe, netMargin, manageCashNetAmount, investCashNetAmount, financingAmount, netAsset, liabilityWithInterestRate))
			//			}
			rowid++
		}
		//归母净利润增长率
		gmNetMarginGrowthRates := p.getGrowthRates(gmNetMargins)
		//总资产增长率
		totalAssetGrowthRates := p.getGrowthRates(totalAssets)
		//归母净资产增长率
		gmNetAssetGrowthRates := p.getGrowthRates(gmNetAssets)
		//营业收入增长率
		openIncomeGrowthRates := p.getGrowthRates(openIncomes)
		//净利润波动率
		fmt.Println(netMarginGrowthRates)
		netMarginVolatilitys := p.getVolatility(netMarginGrowthRates)
		fmt.Println(netMarginVolatilitys)
		//净资产收益率(roe)波动率
		roeVolatilitys := p.getVolatility(roes)
		//投资率
		investGates := p.getInvestGates(investCashNetAmounts, netMargins)
		//拼接进去
		buff := bytes.Buffer{}
		var netMarginVolatility, roeVolatility, investGate, gmNetMarginGrowthRate, totalAssetGrowthRate, gmNetAssetGrowthRate, openIncomeGrowthRate float64
		for rowid, strRow := range strRows {
			if len(netMarginVolatilitys) > rowid {
				netMarginVolatility = netMarginVolatilitys[rowid]
			}
			if len(roeVolatilitys) > rowid {
				roeVolatility = roeVolatilitys[rowid]
			}
			if len(investGates) > rowid {
				investGate = investGates[rowid]
			}
			if len(gmNetMarginGrowthRates) > rowid {
				gmNetMarginGrowthRate = gmNetMarginGrowthRates[rowid]
			}
			if len(totalAssetGrowthRates) > rowid {
				totalAssetGrowthRate = totalAssetGrowthRates[rowid]
			}
			if len(gmNetAssetGrowthRates) > rowid {
				gmNetAssetGrowthRate = gmNetAssetGrowthRates[rowid]
			}
			if len(openIncomeGrowthRates) > rowid {
				openIncomeGrowthRate = openIncomeGrowthRates[rowid]
			}
			if len(netMarginGrowthRates) > rowid {
				netMarginGrowthRate = netMarginGrowthRates[rowid]
			}
			buff.WriteString(fmt.Sprintf("%s|%.2f|%.2f|%.2f|%.2f|%.2f|%.2f|%.2f|%.2f$", strRow, netMarginVolatility, roeVolatility, investGate, gmNetMarginGrowthRate, totalAssetGrowthRate, gmNetAssetGrowthRate, openIncomeGrowthRate, netMarginGrowthRate))
		}
		info = *RemoveLastChar(buff)
		rds.Do("set", key, info)
		rds.Do("expire", key, ExpTime_Short)
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

//获取增长率
func (p *Profitability) getGrowthRates(infos []float64) []float64 {
	growthRates := []float64{}
	for i, info := range infos {
		if i > 0 {
			growthRates = append(growthRates, (info/infos[i-1]-1)*100)
		}
	}
	return growthRates
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

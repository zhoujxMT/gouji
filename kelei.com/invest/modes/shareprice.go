package modes

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/garyburd/redigo/redis"
	"github.com/gin-gonic/gin"

	. "kelei.com/utils/common"
	"kelei.com/utils/frame"
	"kelei.com/utils/logger"
)

type SharePrice struct {
}

type Result_ShowSharePrice struct {
	SharePriceData string
}

func (e *SharePrice) ShowSharePrice(c *gin.Context, args []string) {
	id := args[0]
	//获取投资次数
	enterprise := Enterprise{}
	info := *enterprise.getEnterprise(id)
	if info == "" {
		logger.Errorf("企业不存在")
		return
	}
	sharePriceData := *e.getSharePrice(id)
	res := Result_ShowSharePrice{sharePriceData}
	c.JSON(http.StatusOK, res)
}

//获取企业结构
func (e *SharePrice) getSharePrice(id string) *string {
	rds := frame.GetRedis()
	defer rds.Close()
	key := fmt.Sprintf("shareprice:%s", id)
	info, err := redis.String(rds.Do("get", key))
	if err != nil {
		db := frame.GetDB()
		rows, err := db.Query("select Cycle,RecentChange,SCZSChange from sharepricedata where enterpriseid=?;", id)
		logger.CheckFatal(err, "getSharePrice")
		defer rows.Close()
		buff := bytes.Buffer{}
		var Cycle, RecentChange, SCZSChange string
		for rows.Next() {
			rows.Scan(&Cycle, &RecentChange, &SCZSChange)
			buff.WriteString(fmt.Sprintf("%s|%s|%s$", Cycle, RecentChange, SCZSChange))
		}
		info = *RemoveLastChar(buff)
		rds.Do("set", key, info)
		rds.Do("expire", key, ExpTime_Short)
	}
	return &info
}

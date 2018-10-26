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

type EnterpriseList struct {
}

type Result_GetEnterpriseList struct {
	Enterprises string
}

func (this *EnterpriseList) GetEnterpriseList(c *gin.Context, args []string) {
	rds := frame.GetRedis()
	defer rds.Close()
	key := "EnterpriseList"
	info, err := redis.String(rds.Do("get", key))
	if err != nil {
		db := frame.GetDB()
		rows, err := db.Query("select id,stockName,releaseTime from enterprise")
		logger.CheckFatal(err, "GetEnterpriseList")
		defer rows.Close()
		buff := bytes.Buffer{}
		var id string
		var stockName string
		var releaseTime string
		for rows.Next() {
			rows.Scan(&id, &stockName, &releaseTime)
			t, _ := ParseTime(releaseTime)
			releaseTime = t.Format("2006年1月2日")
			buff.WriteString(fmt.Sprintf("%s$%s$%s|", id, stockName, releaseTime))
		}
		info = *RemoveLastChar(buff)
		rds.Do("set", key, info)
		rds.Do("expire", key, ExpTime_Short)
	}
	res := Result_GetEnterpriseList{info}
	c.JSON(http.StatusOK, res)
}

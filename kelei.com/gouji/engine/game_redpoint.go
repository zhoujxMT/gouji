package engine

import (
	"strconv"

	"kelei.com/utils/logger"
)

//获取红点信息根据类型
func GetRedPoint(args []string) *string {
	userid := args[0]
	redPointType_s := args[1]
	user := UserManage.createUser()
	user.setUserID(&userid)
	redPointType, err := strconv.Atoi(redPointType_s)
	logger.CheckFatal(err, "GetRedPoint")
	return user.GetRedPoint(redPointType)
}

/*
游戏服务器操作-背包
*/

package engine

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"kelei.com/utils/common"
	"kelei.com/utils/logger"
)

const (
	ItemID_FleeReset     = 10022
	ItemID_IntegralReset = 10023
)

/*
获取背包信息
in:页签编号(从0开始)
out:-1没有道具
	baginfoid$itemid$count|baginfoid$itemid$count
*/
func GetBagInfo(args []string) *string {
	userid := args[0]
	user := UserManage.createUser()
	user.setUserID(&userid)

	baginfo_s := user.getBagInfo()
	if *baginfo_s == common.Res_NoData {
		return baginfo_s
	}
	baginfo := strings.Split(*baginfo_s, "|")
	buff := bytes.Buffer{}
	for _, item_s := range baginfo {
		item := strings.Split(item_s, "$")
		iteminfo := common.StrArrToIntArr(item)
		baginfoid := iteminfo[0]
		itemid := iteminfo[1]
		count := iteminfo[2]
		showinbag := iteminfo[4]
		if showinbag == 1 {
			buff.WriteString(fmt.Sprintf("%d$%d$%d|", baginfoid, itemid, count))
		}
	}
	info := common.RowsBufferToString(buff)
	return info
}

/*
逃跑次数清零
in:
out:-2道具不足
	>=0成功,返回道具剩余数量
*/
func FleeReset(args []string) *string {
	res := "-2"
	userid := args[0]
	user := UserManage.createUser()
	user.setUserID(&userid)
	itemid := ItemID_FleeReset
	res = user.useItemByItemID(itemid, 1)
	return &res
}

/*
积分清零
in:
out:-2道具不足
	>=0成功,返回道具剩余数量
*/
func IntegralReset(args []string) *string {
	res := "-2"
	userid := args[0]
	user := UserManage.createUser()
	user.setUserID(&userid)
	itemid := ItemID_IntegralReset
	res = user.useItemByItemID(itemid, 1)
	return &res
}

/*
使用道具(通过baginfoid)
in:baginfoid,数量
out:p1,p2
	p1: -1道具不存在 -2道具不足 -9失败 >=0道具数量
	p2: 使用道具获得的奖励信息
*/
func UseItem(args []string) *string {
	res := "-2"
	userid := args[0]
	baginfoid_s := args[1]
	count_s := args[2]
	baginfoid, err := strconv.Atoi(baginfoid_s)
	logger.CheckFatal(err, "UseItem:1")
	count_int, err2 := strconv.Atoi(count_s)
	logger.CheckFatal(err2, "UseItem:2")
	count := uint(count_int)
	user := UserManage.createUser()
	user.setUserID(&userid)
	res = user.useItem(baginfoid, count)
	return &res
}

/*
使用道具(通过itemid)
in:itemid,数量
out:p1: -1道具不存在 -2道具不足 -9失败 >=0道具数量
	p2: 使用道具获得的奖励信息
*/
func UseItemByItemID(args []string) *string {
	res := "-2"
	userid := args[0]
	itemid_s := args[1]
	count_s := args[2]
	itemid, err := strconv.Atoi(itemid_s)
	logger.CheckFatal(err, "UseItem:1")
	count_int, err2 := strconv.Atoi(count_s)
	logger.CheckFatal(err2, "UseItem:2")
	count := uint(count_int)
	user := UserManage.createUser()
	user.setUserID(&userid)
	res = user.useItemByItemID(itemid, count)
	return &res
}

/*
获取加成道具列表
in:
out:itemid@count#
*/
func GetItemEffect(args []string) *string {
	userid := args[0]
	user := User{}
	user.setUserID(&userid)
	itemEffect := user.GetItemEffect()
	return itemEffect
}

/*
获取比赛中使用的道具
in:
out:-1没有道具
	itemid$count|
*/
func GetMatchItem(args []string) *string {
	userid := args[0]
	user := User{}
	user.setUserID(&userid)
	matchItem := user.GetMatchItem()
	return matchItem
}

/*
玩家数据库操作
*/

package engine

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"strconv"
	"strings"

	"github.com/garyburd/redigo/redis"
	"github.com/mikemintang/go-curl"

	. "kelei.com/utils/common"
	"kelei.com/utils/frame"
	"kelei.com/utils/logger"
	"kelei.com/utils/rpcs"
)

const (
	ItemHandle_Use = iota
	ItemHandle_Add
)

const (
	HonorType_1 = iota + 1
	HonorType_2
	HonorType_3
	HonorType_4
	HonorType_5
	HonorType_6
)

/*
=========================================================================================================
sdk
=========================================================================================================
*/

func (u *User) ComputeHmac256(message string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	return fmt.Sprintf("%x", h.Sum(nil))
}

//上传数据到微信平台
func (u *User) uploadWX() {
	if u.GetSessionKey() == "" {
		return
	}
	userInfo, err := redis.Values(u.GetUserInfo("win", "charm"))
	logger.CheckFatal(err, "reportedUserInfo")
	win, _ := redis.String(userInfo[0], nil)
	charm, _ := redis.String(userInfo[1], nil)

	postData := map[string]interface{}{
		"kv_list": []map[string]string{
			map[string]string{
				"key":   "win",
				"value": win,
			}, map[string]string{
				"key":   "charm",
				"value": charm,
			},
		},
	}
	str := `{"kv_list":[{"key":"win","value":"` + win + `"},{"key":"charm","value":"` + charm + `"}]}`
	signature := u.ComputeHmac256(str, u.GetSessionKey())
	postUrl := fmt.Sprintf("https://api.weixin.qq.com/wxa/set_user_storage?access_token=%s&signature=%s&openid=%s&sig_method=%s", u.GetToken(), signature, *u.getUID(), "hmac_sha256")
	req := curl.NewRequest()
	req.SetDialTimeOut(2)
	resp, err := req.SetUrl(postUrl).SetPostData(postData).Post()
	if err != nil {
		logger.Errorf(err.Error())
	} else {
		if resp.IsOk() {
			fmt.Println(resp.Body)
		} else {
			fmt.Println(resp.Raw)
		}
	}
}

/*
=========================================================================================================
推送
=========================================================================================================
*/

//给玩家推送提示条(屏幕上方下拉的方式弹出“提示条”)
func (u *User) topBar(title, info string) {
	message := title + "|" + info
	u.push("TopBar_Push", &message)
}

/*
=========================================================================================================
负载均衡服务器
=========================================================================================================
*/

//向负载均衡服务器插入此玩家的比赛信息，用来重新进入
func (u *User) insertUserInfo() {
	userid := *u.getUserID()
	room := u.getRoom()
	roomid := room.id
	matchid := strconv.Itoa(room.GetMatchID())
	roomtype := strconv.Itoa(room.GetRoomType())
	xclient := frame.NewRpcxClient(frame.GetArgs().RpcxClient)
	defer xclient.Close()
	args := &rpcs.Args{}
	args.V = map[string]interface{}{"args": []string{userid, addr, roomid, matchid, roomtype}}
	reply := &rpcs.Reply{}
	err := xclient.Call(context.Background(), "InsertUserInfo", args, reply)
	logger.CheckError(err, "insertUserInfo")
}

//从负载均衡服务器删除此玩家的比赛信息
func (u *User) deleteUserInfo() {
	userid := *u.getUserID()
	xclient := frame.NewRpcxClient(frame.GetArgs().RpcxClient)
	defer xclient.Close()
	args := &rpcs.Args{}
	args.V = map[string]interface{}{"args": []string{userid}}
	reply := &rpcs.Reply{}
	err := xclient.Call(context.Background(), "DeleteUserInfo", args, reply)
	logger.CheckError(err, "deleteUserInfo")
}

/*
=========================================================================================================
登录服务器
=========================================================================================================
*/

//获取nmmember-redis中是否存在玩家信息
func (u *User) memberInfoExists() bool {
	nmRds := getNMRds()
	defer nmRds.Close()
	exists, err := redis.Int(nmRds.Do("exists", u.getKey()))
	logger.CheckFatal(err, "memberInfoExists")
	return exists == 1
}

//加载memberinfo信息到nmmember-redis中
func (u *User) loadMemberInfo() {
	if !u.memberInfoExists() {
		heroCoin := 0
		userInfo, _ := redis.Strings(u.GetUserInfo("uid"))
		u.setUID(&userInfo[0])
		err := nmDB.QueryRow("select heroCoin from memberinfo where uid=?", *u.getUID()).Scan(&heroCoin)
		logger.CheckFatal(err, "loadMemberInfo")
		nmRds := getNMRds()
		defer nmRds.Close()
		nmRds.Do("hmset", u.getKey(), "heroCoin", heroCoin)
		nmRds.Do("expire", u.getKey(), expireSecond)
	}
}

//获取英雄币数量
func (u *User) getHeroCoin() int {
	u.loadMemberInfo()
	nmRds := getNMRds()
	defer nmRds.Close()
	heroCoin, err := redis.Int(nmRds.Do("hget", u.getKey(), "heroCoin"))
	logger.CheckFatal(err, "getHeroCoin")
	return heroCoin
}

/*
更新英雄币数量
in:变动的英雄币
out:剩余的英雄币
*/
func (u *User) updateHeroCoin(value int, src int) int {
	//执行存储过程
	heroCoin := 0
	userInfo, _ := redis.Strings(u.GetUserInfo("uid"))
	u.setUID(&userInfo[0])
	err := nmDB.QueryRow("call HeroCoinUpdate(?,?,?)", *u.getUID(), value, src).Scan(&heroCoin)
	logger.CheckFatal(err, "updateHeroCoin")
	nmRds := getNMRds()
	defer nmRds.Close()
	//设置键过期
	nmRds.Do("expire", u.getKey(), 0)
	return heroCoin
}

/*
获取member信息
*/
func (u *User) GetMemberInfo(args ...interface{}) (interface{}, error) {
	u.loadMemberInfo()
	args = append(args[:0], append([]interface{}{u.getKey()}, args[0:]...)...)
	nmRds := getNMRds()
	defer nmRds.Close()
	return nmRds.Do("hmget", args...)
}

/*
=========================================================================================================
游戏服务器
=========================================================================================================
*/

//获取game-redis中是否存在玩家信息
func (u *User) userInfoExists() bool {
	gameRds := getGameRds()
	defer gameRds.Close()
	exists, err := redis.Int(gameRds.Do("exists", u.getKey()))
	logger.CheckFatal(err, "userInfoExists")
	return exists == 1
}

//获取真实的userid从AI的userid中
func (u *User) getRealityUserID() string {
	userid := *u.getUserID()
	if len(userid) > 13 {
		userid = userid[13:]
	}
	return userid
}

//加载userinfo信息到game-redis中
func (u *User) loadUserInfo() {
	if !u.userInfoExists() {
		username, uid := "", ""
		ingot, sex, level, integral, lvlExp, upExp, exp, charm, flee, inning, win, flat, lose, vip := 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0

		userid := u.getRealityUserID()
		row := db.QueryRow("select username, uid, ingot, sex, level, integral, lvlExp, upExp, exp, charm, flee, inning, win, flat, lose, vip from userinfo where userid=?", userid)
		err := row.Scan(&username, &uid, &ingot, &sex, &level, &integral, &lvlExp, &upExp, &exp, &charm, &flee, &inning, &win, &flat, &lose, &vip)
		logger.CheckFatal(err, fmt.Sprintf("loadUserInfo : userid : ", *u.getUserID()))
		gameRds := getGameRds()
		defer gameRds.Close()
		gameRds.Do("hmset", u.getKey(), "username", username, "uid", uid, "ingot", ingot, "sex", sex, "level", level, "integral", integral, "lvlExp", lvlExp, "upExp", upExp, "exp", exp, "charm", charm, "flee", flee, "inning", inning, "win", win, "flat", flat, "lose", lose, "vip", vip)
		gameRds.Do("expire", u.getKey(), expireSecond_Short)
	}
}

//获取元宝数量
func (u *User) getIngot() int {
	u.loadUserInfo()
	userInfo, _ := redis.Ints(u.GetUserInfo("ingot"))
	ingot := userInfo[0]
	return ingot
}

/*
获取玩家等级
*/
func (u *User) getLevel() int {
	u.loadUserInfo()
	userInfo, _ := redis.Ints(u.GetUserInfo("level"))
	level := userInfo[0]
	return level
}

/*
更新元宝数量
in:变动的元宝
out:剩余的元宝
*/
func (u *User) updateIngot(value int, src int) int {

	//执行存储过程
	ingot := 0
	err := db.QueryRow("call IngotUpdate(?,?,?,?)", Procedure_Return, *u.getUserID(), value, src).Scan(&ingot)
	logger.CheckFatal(err, "updateIngot")
	gameRds := getGameRds()
	defer gameRds.Close()
	//设置键过期
	gameRds.Do("expire", u.getKey(), 0)
	return ingot
}

/*
更新玩家经验
out:是否升级、当前级别、当前级别经验、当前级别升级经验
*/
func (u *User) updateExp(value int) (int, int, int, int, int) {

	//执行存储过程
	isUpgrade, level, lvlExp, upExp, upIngotAward := 0, 0, 0, 0, 0
	err := db.QueryRow("call ExpUpdate(1,?,?)", *u.getUserID(), value).Scan(&isUpgrade, &level, &lvlExp, &upExp, &upIngotAward)
	logger.CheckFatal(err, "updateExp")
	gameRds := getGameRds()
	defer gameRds.Close()
	//设置键过期
	gameRds.Do("expire", u.getKey(), 0)
	return isUpgrade, level, lvlExp, upExp, upIngotAward
}

/*
更新玩家积分
out:是否升级、当前级别、当前级别经验、当前级别升级经验
*/
func (u *User) updateIntegral(value int) {
	//积分清零
	if value == 0 {
		_, err := db.Exec("update UserInfo set integral=0 where UserInfo.UserID=?;", *u.getUserID())
		logger.CheckFatal(err, "updateIntegral")
	} else {
		_, err := db.Exec("update UserInfo set integral=integral+? where UserInfo.UserID=?;", value, *u.getUserID())
		logger.CheckFatal(err, "updateIntegral2")
	}
	gameRds := getGameRds()
	defer gameRds.Close()
	//设置键过期
	gameRds.Do("expire", u.getKey(), 0)
}

/*
更新玩家魅力值
out:是否升级、当前级别、当前级别经验、当前级别升级经验
*/
func (u *User) updateCharm(value int) {
	db.QueryRow("call CharmUpdate(0,?,?)", *u.getUserID(), value)
	gameRds := getGameRds()
	defer gameRds.Close()
	//设置键过期
	gameRds.Do("expire", u.getKey(), 0)
}

/*
	魅力值清零
*/
func (u *User) resetCharm() {
	//	charm, _ := u.GetUserInfo("charm")
}

/*
获取玩家逃跑的费用
in:
out:元宝|积分
*/
func (u *User) FleeCost() string {
	room := u.getRoom()
	roomData, err := redis.Ints(room.GetRoomData("fleeIngot", "fleeIntegral"))
	logger.CheckFatal(err)
	fleeIngot, fleeIntegral := roomData[0], roomData[1]
	res := fmt.Sprintf("%d|%d", fleeIngot, fleeIntegral)
	return res
}

/*
更新玩家逃跑次数
*/
func (u *User) updateFlee(value int) {
	//逃跑次数清零
	if value == 0 {
		_, err := db.Exec("update UserInfo set flee=0 where UserInfo.UserID=?;", *u.getUserID())
		logger.CheckFatal(err, "updateFlee")
	} else {
		_, err := db.Exec("update UserInfo set flee=flee+? where UserInfo.UserID=?;", value, *u.getUserID())
		logger.CheckFatal(err, "updateFlee2")
	}
	gameRds := getGameRds()
	defer gameRds.Close()
	//设置键过期
	gameRds.Do("expire", u.getKey(), 0)
}

/*
开赛前-获取道具效果信息
*/
func (u *User) beginGetItemEffect() {
	//保存道具效果信息
	u.setItemEffect(*u.GetItemEffect())
	//保存道具类型信息

	itemsEffectInfo := ""
	err := db.QueryRow("call GetItemEffect(?)", *u.getUserID()).Scan(&itemsEffectInfo)
	logger.CheckFatal(err, "beginGetItemEffect")
	//绑定玩家身上的道具效果
	if itemsEffectInfo != "" {
		itemsEffectInfo = itemsEffectInfo[:len(itemsEffectInfo)-1]
		itemsEffect := strings.Split(itemsEffectInfo, "|")
		effectTypes := map[int]int{}
		for _, itemEffect_s := range itemsEffect {
			itemEffect := strings.Split(itemEffect_s, "$")
			effectType, _ := strconv.Atoi(itemEffect[0])
			effectVal, _ := strconv.Atoi(itemEffect[1])
			effectTypes[effectType] = effectVal
		}
		u.setEffectTypes(effectTypes)
	} else {
		u.setEffectTypes(nil)
	}
}

/*
开赛-更新玩家数据
in:积分、魅力
*/
func (u *User) beginUpdateUserInfo(mapInfo map[string]int) {
	integral := mapInfo["integral"]
	charm := mapInfo["charm"]

	charm = u.effect_Charm(charm)
	_, err := db.Exec("call MatchBegan(?,?,?)", *u.getUserID(), integral, charm)
	logger.CheckFatal(err, "beginUpdateUserInfo")
	gameRds := getGameRds()
	defer gameRds.Close()
	//设置键过期
	gameRds.Do("expire", u.getKey(), 0)
}

/*
比赛结束-更新玩家数据
in:积分、MatchResult
*/
func (u *User) endUpdateUserInfo(mapInfo map[string]int) {
	integral := mapInfo["integral"]
	matchResult := mapInfo["matchResult"]

	win, flat, lose := 0, 0, 0
	if matchResult == MatchResult_Win {
		win = 1
	} else if matchResult == MatchResult_Flat {
		flat = 1
	} else if matchResult == MatchResult_Lose {
		lose = 1
	}
	roomType := u.getRoom().GetRoomType()
	_, err := db.Exec("call MatchEnd(?,?,?,?,?,?)", *u.getUserID(), roomType, integral, win, flat, lose)
	logger.CheckFatal(err, "endUpdateUserInfo")
	gameRds := getGameRds()
	defer gameRds.Close()
	//设置键过期
	gameRds.Do("expire", u.getKey(), 0)
	gameRds.Do("expire", u.getHonorInfoKey(HonorType_1), 0)
	gameRds.Do("expire", u.getHonorInfoKey(HonorType_2), 0)
	gameRds.Do("expire", u.getHonorInfoKey(HonorType_3), 0)
	gameRds.Do("expire", u.getHonorInfoKey(HonorType_4), 0)
	gameRds.Do("expire", u.getHonorInfoKey(HonorType_6), 0)
}

/*
获取玩家信息
*/
func (u *User) GetUserInfo(args ...interface{}) (interface{}, error) {
	u.loadUserInfo()
	args = append(args[:0], append([]interface{}{u.getKey()}, args[0:]...)...)
	gameRds := getGameRds()
	defer gameRds.Close()
	return gameRds.Do("hmget", args...)
}

/*
------------------------------------------------------------------------------------------------------------
背包
------------------------------------------------------------------------------------------------------------
*/

//获取baginfo的key
func (u *User) getBagInfoKey() string {
	return "baginfo:" + *u.getUserID()
}

//加载baginfo信息到game-redis中
func (u *User) getBagInfo() *string {
	gameRds := getGameRds()
	defer gameRds.Close()
	info, err := redis.String(gameRds.Do("get", u.getBagInfoKey()))
	if err != nil {

		rows, err := db.Query("select baginfoid,baginfo.itemid,count,itemdata.effectType,itemdata.showinbag from baginfo left join itemdata on baginfo.itemid=itemdata.itemid where baginfo.userid=?;", *u.getUserID())
		logger.CheckFatal(err, "getBagInfo")
		defer rows.Close()
		buff := bytes.Buffer{}
		baginfoid, itemid, count, effectType := 0, 0, 0, 0
		showInBag := []uint8{}
		for rows.Next() {
			rows.Scan(&baginfoid, &itemid, &count, &effectType, &showInBag)
			buff.WriteString(fmt.Sprintf("%d$%d$%d$%d$%d|", baginfoid, itemid, count, effectType, int(showInBag[0])))
		}
		info = *RowsBufferToString(buff)
		gameRds.Do("set", u.getBagInfoKey(), info)
		gameRds.Do("expire", u.getBagInfoKey(), expireSecond)
	}
	return &info
}

//获取AdditionItems的key
func (u *User) getItemEffectKey() string {
	return "itemEffect:" + *u.getUserID()
}

//获取加成道具信息
func (u *User) GetItemEffect() *string {
	gameRds := getGameRds()
	defer gameRds.Close()
	key := u.getItemEffectKey()
	info, err := redis.String(gameRds.Do("get", key))
	if err != nil {

		rows, err := db.Query("select itemid,val from ItemEffect where userid=?;", *u.getUserID())
		logger.CheckFatal(err, "GetItemEffect")
		defer rows.Close()
		buff := bytes.Buffer{}
		itemid, val := 0, 0
		for rows.Next() {
			rows.Scan(&itemid, &val)
			buff.WriteString(fmt.Sprintf("%d@%d#", itemid, val))
		}
		info = *RowsBufferToString(buff)
		gameRds.Do("set", key, info)
		gameRds.Do("expire", key, expireSecond)
	}
	return &info
}

//获取比赛中的道具
func (u *User) GetMatchItem() *string {
	baginfo_s := u.getBagInfo()
	if *baginfo_s == Res_NoData {
		return baginfo_s
	}
	baginfo := strings.Split(*baginfo_s, "|")
	buff := bytes.Buffer{}
	for _, item_s := range baginfo {
		item := strings.Split(item_s, "$")
		iteminfo := StrArrToIntArr(item)
		itemid := iteminfo[1]
		count := iteminfo[2]
		effectType := iteminfo[3]
		if effectType == EffectType_MatchItem {
			buff.WriteString(fmt.Sprintf("%d$%d|", itemid, count))
		}
	}
	info := RowsBufferToString(buff)
	return info
}

//获取玩家道具信息根据BagInfoID
func (u *User) getItemByBagInfoID(bagInfoID int) *Item {
	item := &Item{bagInfoID, 0, 0}
	bagInfo := *u.getBagInfo()
	if bagInfo == "" {
		return item
	}
	arrItemInfo := strings.Split(bagInfo, "|")
	for _, itemInfo := range arrItemInfo {
		item_s := strings.Split(itemInfo, "$")
		item_i := StrArrToIntArr(item_s)
		baginfoid_, itemid_, count_ := item_i[0], item_i[1], item_i[2]
		if baginfoid_ == bagInfoID {
			item = &Item{baginfoid_, itemid_, count_}
			break
		}
	}
	return item
}

//获取玩家道具信息根据ItemID
func (u *User) getItemByItemID(itemid int) *Item {
	item := &Item{0, itemid, 0}
	bagInfo := *u.getBagInfo()
	if bagInfo == "" {
		return item
	}
	arrItemInfo := strings.Split(bagInfo, "|")
	for _, itemInfo := range arrItemInfo {
		item_s := strings.Split(itemInfo, "$")
		item_i := StrArrToIntArr(item_s)
		baginfoid_, itemid_, count_ := item_i[0], item_i[1], item_i[2]
		if itemid_ == itemid {
			item = &Item{baginfoid_, itemid_, count_}
			break
		}
	}
	return item
}

/*
通过baginfoid使用道具
in:baginfoid,道具数量
out:-1道具不存在 -2道具不足 -9失败 >=0道具数量
*/
func (u *User) useItem(baginfoid int, count uint, args ...string) string {
	info := ""
	if len(args) > 0 {
		info = args[0]
	}

	//执行存储过程
	res := 0
	res_iteminfo := ""
	err := db.QueryRow("call ItemHandle(?,?,?,?,?,?,?)", Procedure_Return, ItemHandle_Use, *u.getUserID(), baginfoid, count, 1, info).Scan(&res, &res_iteminfo)
	logger.CheckFatal(err, "useItem")
	gameRds := getGameRds()
	defer gameRds.Close()
	//设置键过期
	gameRds.Do("expire", u.getKey(), 0)
	gameRds.Do("expire", u.getBagInfoKey(), 0)
	gameRds.Do("expire", u.getHonorInfoKey(HonorType_5), 0)
	return fmt.Sprintf("%d,%s", res, res_iteminfo)
}

/*
通过itemid使用道具
in:itemid,道具数量
out:-1道具不存在 -2道具不足 -9失败 >=0道具数量
*/
func (u *User) useItemByItemID(itemid int, count uint, args ...string) string {
	res := "-2"
	item := u.getItemByItemID(itemid)
	userCount := item.getCount()
	//道具不足
	if userCount <= 0 {
		return res
	}
	res = u.useItem(item.getBagInfoID(), count, args...)
	return res
}

/*
添加道具
in:itemid,道具数量
out:-1道具不存在 -9失败
	>=0道具数量
*/
func (u *User) addItem(itemid int, count uint) int {

	//执行存储过程
	res := 0
	err := db.QueryRow("call ItemHandle(?,?,?,?,?,?,?)", Procedure_Return, ItemHandle_Add, *u.getUserID(), itemid, count, 2, "").Scan(&res)
	logger.CheckFatal(err, "addItem")
	gameRds := getGameRds()
	defer gameRds.Close()
	//设置键过期
	gameRds.Do("expire", u.getBagInfoKey(), 0)
	return res
}

/*
------------------------------------------------------------------------------------------------------------
成就
------------------------------------------------------------------------------------------------------------
*/

//获取baginfo的key
func (u *User) getHonorInfoKey(honorType int) string {
	return fmt.Sprintf("honorinfo:%s:%d", *u.getUserID(), honorType)
}

/*
获得成就信息
in:成就类型
out:-1没有成就信息
	id$AchievementDataID$当前数量$状态(0未达成1已达成2已领奖)|
*/
func (u *User) GetHonorInfo(honorType int) *string {
	gameRds := getGameRds()
	defer gameRds.Close()
	key := u.getHonorInfoKey(honorType)
	info, err := redis.String(gameRds.Do("get", key))
	if err != nil {

		rows, err := db.Query("select AchievementInfo.id,AchievementInfo.AchievementDataID,AchievementInfo.nowcount,AchievementInfo.status from AchievementInfo left join AchievementData on AchievementInfo.AchievementDataID=AchievementData.ID where AchievementInfo.userid=? and AchievementData.type=?;", *u.getUserID(), honorType)
		logger.CheckFatal(err, "GetHonorInfo")
		defer rows.Close()
		buff := bytes.Buffer{}
		id, AchievementDataID, nowcount, status := 0, 0, 0, 0
		for rows.Next() {
			rows.Scan(&id, &AchievementDataID, &nowcount, &status)
			buff.WriteString(fmt.Sprintf("%d$%d$%d$%d|", id, AchievementDataID, nowcount, status))
		}
		info = *RowsBufferToString(buff)
		gameRds.Do("set", key, info)
		gameRds.Do("expire", key, expireSecond_Long_Long)
	}
	return &info
}

/*
成就领奖
in:id
out:-1成就信息不存在 -2成就未达成 -3成就已领奖
	1领奖成功
*/
func (u *User) HonorAward(achievementInfoID int) *string {

	//执行存储过程
	res := "1"
	err2 := db.QueryRow("call HonorAward(?,?)", *u.getUserID(), achievementInfoID).Scan(&res)
	logger.CheckFatal(err2, "HonorAward2")
	res_int, err := strconv.Atoi(res)
	logger.CheckFatal(err, "HonorAward")
	//领奖失败
	if res_int < 0 {
		return &res
	}
	honorType := res_int
	gameRds := getGameRds()
	defer gameRds.Close()
	//设置键过期
	gameRds.Do("expire", u.getBagInfoKey(), 0)
	gameRds.Do("expire", u.getHonorInfoKey(honorType), 0)
	gameRds.Do("expire", u.getKey(), 0)
	res = "1"
	return &res
}

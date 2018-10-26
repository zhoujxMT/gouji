/*
游戏服务器操作-玩家
*/

package engine

import (
	"fmt"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"

	"kelei.com/utils/logger"

	. "kelei.com/utils/common"
)

/*
创建玩家,如果玩家不存在就创建
in:
out:userid
*/
func CreateUser(args []string) *string {
	res := Res_Unknown
	uid := args[0]
	username := args[1]

	userid := -1
	db.QueryRow("select userinfo.userid from userinfo where userinfo.uid=?", uid).Scan(&userid)
	if userid == -1 {
		_, err := db.Exec("insert into userinfo(uid,username) values(?,?)", uid, username)
		logger.CheckFatal(err, "CreateUser2")
		err = db.QueryRow("select userid from userinfo where uid=?", uid).Scan(&userid)
		logger.CheckFatal(err, "CreateUser3")
		_, err = db.Exec("call UserInit(?)", userid)
		logger.CheckFatal(err, "CreateUser4")
	} else {
		db.Exec("update userinfo set userinfo.username=? where uid=?", username, uid)
	}
	res = strconv.Itoa(userid)
	return &res
}

/*
获取玩家信息
in:userid
out:-2获取信息失败
	等级,当前级别经验,当前级别升级经验,魅力值,积分,元宝数,逃跑次数,总局数,胜局,平局,败局,userid
*/
func GetUserInfo(args []string) *string {
	res := "-2"
	userid := args[1]
	//创建玩家
	user := UserManage.createUser()
	user.setUserID(&userid)
	//获取元宝
	userIngot := user.getIngot()
	userIngot = userIngot
	//获取玩家信息
	userInfo, err := redis.Ints(user.GetUserInfo("level", "lvlExp", "upExp", "charm", "integral", "flee", "inning", "win", "flat", "lose", "vip"))
	logger.CheckError(err)
	level, lvlExp, upExp, charm, integral, flee, inning, win, flat, lose, vip := userInfo[0], userInfo[1], userInfo[2], userInfo[3], userInfo[4], userInfo[5], userInfo[6], userInfo[7], userInfo[8], userInfo[9], userInfo[10]
	res = fmt.Sprintf("%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%s,%d", level, lvlExp, upExp, charm, integral, userIngot, flee, inning, win, flat, lose, userid, vip)
	return &res
}

/*
设置玩家托管状态
in:托管状态(0不拖1拖)
out:没有返回值
push:玩家托管状态
*/
func SetUserTG(args []string) *string {
	res := Res_NoBack
	userid := args[0]
	status := args[1]
	user := UserManage.GetUser(&userid)
	if user == nil {
		res = Res_Unknown
		return &res
	}
	user.setTrusteeship(status == "1")
	return &res
}

/*
获取逃跑的费用（不在房间中返回未知数据）
in:
out:元宝|积分
*/
func FleeCost(args []string) *string {
	res := Res_Unknown
	userid := args[0]
	user := UserManage.GetUser(&userid)
	//不在比赛中
	if user == nil {
		return &res
	}
	room := user.getRoom()
	if room.GetRoomStatus() == RoomStatus_Setout {
		return &res
	}
	res = user.FleeCost()
	return &res
}

/*
发送表情和文字（只有玩家在房间中的时候才有返回值）
in:信息类型,信息编号
out:无返回值
push:userid,信息类型,信息编号
*/
func Chat(args []string) *string {
	res := Res_NoBack
	userid := args[0]
	user := UserManage.GetUser(&userid)
	if user == nil || user.getChatLastTime().Add(time.Second*2).After(time.Now()) {
		return &res
	}
	user.setChatLastTime()
	//不在房间中
	if user == nil {
		return &res
	}
	//没落座
	if user.getStatus() == UserStatus_NoSitDown {
		return &res
	}
	contentType := args[1]
	contentIndex := args[2]
	room := user.getRoom()
	pushMessageToUsers("Chat_Push", []string{fmt.Sprintf("%s|%s|%s", userid, contentType, contentIndex)}, room.getAllUserIDs())
	return &res
}

//获得成就信息
func GetHonorInfo(args []string) *string {
	userid := args[0]
	honorType_s := args[1]
	user := UserManage.createUser()
	user.setUserID(&userid)
	honorType, err := strconv.Atoi(honorType_s)
	logger.CheckFatal(err)
	return user.GetHonorInfo(honorType)
}

//成就领奖
func HonorAward(args []string) *string {
	userid := args[0]
	achievementInfoID_s := args[1]
	user := UserManage.createUser()
	user.setUserID(&userid)
	achievementInfoID, err := strconv.Atoi(achievementInfoID_s)
	logger.CheckFatal(err)
	return user.HonorAward(achievementInfoID)
}

/*
提示联邦出牌
in:userid,cardindex|cardindex
out:1提示成功
push:[HintFederal_Push] userid,cardindex|cardindex
*/
func HintFederal(args []string) *string {
	res := "-1"
	userid := args[0]
	federalUserID := args[1]
	cardsIndex := args[2]
	user := UserManage.GetUser(&userid)
	user.HintFederalPlayCard(federalUserID, cardsIndex)
	res = "1"
	return &res
}

/*
获取邮件列表
*/
func GetMails(args []string) *string {
	userid := args[0]
	user := UserManage.createUser()
	user.setUserID(&userid)
	res := user.GetMails()
	return res
}

/*
断线重连获取比赛信息
in:
out:-1房间不存在
	1房间存在
push:
	1.房间存在时{
		1.未开赛只推送“matchingPush”
		2.已开赛推送各种信息除了“matchingPush”
	}
*/
func Reconnect(args []string) *string {
	res := ""
	userid := args[0]
	user := UserManage.GetUser(&userid)
	if user == nil {
		res = "-1"
		return &res
	}
	user.Reconnect()
	res = "1"
	return &res
}

//读取邮件
func ReadMail(args []string) *string {
	userid := args[0]
	mailid, err := strconv.Atoi(args[1])
	logger.CheckFatal(err)
	user := UserManage.newUser(&userid)
	res := user.ReadMail(mailid)
	return res
}

//领取邮件奖励
func ReceMail(args []string) *string {
	userid := args[0]
	mailid, err := strconv.Atoi(args[1])
	logger.CheckFatal(err)
	user := UserManage.newUser(&userid)
	res := user.ReceMail(mailid)
	return res
}

/*
走科或革命,选一家联邦看牌
in:要看牌的玩家的userid
out:1成功 其它失败
*/
func LookCard(args []string) *string {
	userid := args[0]
	lookeduserid := args[1]
	user := UserManage.GetUser(&userid)
	lookeduser := UserManage.GetUser(&lookeduserid)
	res := user.LookCard(lookeduser)
	return res
}

/*
比赛中对其它玩家使用道具
in:目标玩家的userid,itemid
out:道具剩余数量,itemid
push:MUI_Push	out:使用道具的UserID,接受道具的UserID,itemid
des:道具剩余数量>=0 使用成功   <0 使用失败
*/
func MatchUseItem(args []string) *string {
	userid := args[0]
	dstuserid := args[1]
	itemid, err := strconv.Atoi(args[2])
	logger.CheckFatal(err)
	user := UserManage.GetUser(&userid)
	if user == nil {
		return &Res_Unknown
	}
	res := user.MatchUseItem(dstuserid, itemid)
	return res
}

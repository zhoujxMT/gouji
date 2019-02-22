package engine

import (
	"net"
	"strconv"

	"kelei.com/utils/common"
	"kelei.com/utils/logger"
	"kelei.com/utils/rpcs"
)

type Engine struct {
}

func NewEngine(addr string) *Engine {
	logger.Infof("[启动引擎]")
	engine := Engine{}
	db_Init()
	redis_Init()
	engine_Init(addr)
	return &engine
}

func (this *Engine) getArgs(args *rpcs.Args) []string {
	p := args.V["args"].([]interface{})
	newP := []string{}
	for _, arg := range p {
		newP = append(newP, arg.(string))
	}
	return newP
}

func (this *Engine) getUserID(args *rpcs.Args) string {
	p := args.V["args"].([]interface{})
	userid := p[0].(string)
	return userid
}

func (this *Engine) getUser(args *rpcs.Args) *User {
	userid := this.getUserID(args)
	user := UserManage.GetUser(&userid)
	return user
}

/*
	建立连接
*/
func (this *Engine) Connect(args *rpcs.Args) *rpcs.Reply {
	p := this.getArgs(args)
	userid, uid, token, sessionKey, secret, headUrl := p[0], p[1], p[2], p[3], p[4], p[5]
	user := UserManage.AddUser(&uid, &userid, args.V["clientConn"].(net.Conn))
	user.SetToken(token)
	user.SetSessionKey(sessionKey)
	user.SetSecret(secret)
	user.SetHeadUrl(headUrl)
	//更新数据库玩家头像
	db.Exec("update userinfo set headurl=? where userinfo.userid=?", headUrl, userid)
	return &rpcs.Reply{nil, common.SC_OK}
}

/*
	关闭连接
*/
func (this *Engine) ClientConnClose(args *rpcs.Args) *rpcs.Reply {
	res := ClientConnClose(this.getArgs(args))
	return &rpcs.Reply{res, common.SC_OK}
}

/*
	匹配房间
*/
func (this *Engine) Matching(args *rpcs.Args) *rpcs.Reply {
	res := Matching(this.getArgs(args))
	return &rpcs.Reply{res, common.SC_OK}
}

/*
	进入房间
*/
func (this *Engine) EnterRoom(args *rpcs.Args) *rpcs.Reply {
	p := this.getArgs(args)
	p = common.InsertStringSlice(p, []string{strconv.Itoa(MatchingMode_Normal)}, 3)
	res := EnterRoom(p)
	return &rpcs.Reply{res, common.SC_OK}
}

//换桌
func (this *Engine) ChangeTable(args *rpcs.Args) *rpcs.Reply {
	res := ChangeTable(this.getArgs(args))
	return &rpcs.Reply{res, common.SC_OK}
}

//准备
func (this *Engine) Setout(args *rpcs.Args) *rpcs.Reply {
	res := Setout(this.getArgs(args))
	return &rpcs.Reply{res, common.SC_OK}
}

//获取座位状态的推送
func (this *Engine) GetMatchingPush(args *rpcs.Args) *rpcs.Reply {
	res := GetMatchingPush(this.getArgs(args))
	return &rpcs.Reply{res, common.SC_OK}
}

//出牌
func (this *Engine) PlayCard(args *rpcs.Args) *rpcs.Reply {
	res := PlayCard(this.getArgs(args))
	return &rpcs.Reply{res, common.SC_OK}
}

//提示
func (this *Engine) Hint(args *rpcs.Args) *rpcs.Reply {
	res := Hint(this.getArgs(args))
	return &rpcs.Reply{res, common.SC_OK}
}

//过牌
func (this *Engine) CheckCard(args *rpcs.Args) *rpcs.Reply {
	res := CheckCard(this.getArgs(args))
	return &rpcs.Reply{res, common.SC_OK}
}

//强制过牌
func (this *Engine) ForceCheckCard(args *rpcs.Args) *rpcs.Reply {
	res := ForceCheckCard(this.getArgs(args))
	return &rpcs.Reply{res, common.SC_OK}
}

//让牌
func (this *Engine) LetCard(args *rpcs.Args) *rpcs.Reply {
	res := LetCard(this.getArgs(args))
	return &rpcs.Reply{res, common.SC_OK}
}

//烧牌
func (this *Engine) Burn(args *rpcs.Args) *rpcs.Reply {
	res := Burn(this.getArgs(args))
	return &rpcs.Reply{res, common.SC_OK}
}

//不烧牌
func (this *Engine) NoBurn(args *rpcs.Args) *rpcs.Reply {
	res := NoBurn(this.getArgs(args))
	return &rpcs.Reply{res, common.SC_OK}
}

//设置玩家托管
func (this *Engine) SetUserTG(args *rpcs.Args) *rpcs.Reply {
	res := SetUserTG(this.getArgs(args))
	return &rpcs.Reply{res, common.SC_OK}
}

//退出比赛
func (this *Engine) ExitMatch(args *rpcs.Args) *rpcs.Reply {
	res := ExitMatch(this.getArgs(args))
	return &rpcs.Reply{res, common.SC_OK}
}

//获取主场景信息
func (this *Engine) GetMainInfo(args *rpcs.Args) *rpcs.Reply {
	res := GetMainInfo(this.getArgs(args))
	return &rpcs.Reply{res, common.SC_OK}
}

//进入某类型的比赛
func (this *Engine) EnterMatch(args *rpcs.Args) *rpcs.Reply {
	res := EnterMatch(this.getArgs(args))
	return &rpcs.Reply{res, common.SC_OK}
}

//获取比赛类型下的房间类型信息
func (this *Engine) GetMatchRoomInfo(args *rpcs.Args) *rpcs.Reply {
	res := GetMatchRoomInfo(this.getArgs(args))
	return &rpcs.Reply{res, common.SC_OK}
}

//获取玩家信息
func (this *Engine) GetUserInfo(args *rpcs.Args) *rpcs.Reply {
	res := GetUserInfo(this.getArgs(args))
	return &rpcs.Reply{res, common.SC_OK}
}

//获取背包信息
func (this *Engine) GetBagInfo(args *rpcs.Args) *rpcs.Reply {
	res := GetBagInfo(this.getArgs(args))
	return &rpcs.Reply{res, common.SC_OK}
}

//积分清零
func (this *Engine) IntegralReset(args *rpcs.Args) *rpcs.Reply {
	res := IntegralReset(this.getArgs(args))
	return &rpcs.Reply{res, common.SC_OK}
}

//逃跑次数清零
func (this *Engine) FleeReset(args *rpcs.Args) *rpcs.Reply {
	res := FleeReset(this.getArgs(args))
	return &rpcs.Reply{res, common.SC_OK}
}

//获取房间逃跑的费用
func (this *Engine) FleeCost(args *rpcs.Args) *rpcs.Reply {
	res := FleeCost(this.getArgs(args))
	return &rpcs.Reply{res, common.SC_OK}
}

//使用道具(通过baginfoid)
func (this *Engine) UseItem(args *rpcs.Args) *rpcs.Reply {
	res := UseItem(this.getArgs(args))
	return &rpcs.Reply{res, common.SC_OK}
}

//使用道具(通过itemid)
func (this *Engine) UseItemByItemID(args *rpcs.Args) *rpcs.Reply {
	res := UseItemByItemID(this.getArgs(args))
	return &rpcs.Reply{res, common.SC_OK}
}

//发送表情和文字
func (this *Engine) Chat(args *rpcs.Args) *rpcs.Reply {
	res := Chat(this.getArgs(args))
	return &rpcs.Reply{res, common.SC_OK}
}

//获取充值信息
func (this *Engine) GetPayInfo(args *rpcs.Args) *rpcs.Reply {
	res := GetPayInfo(this.getArgs(args))
	return &rpcs.Reply{res, common.SC_OK}
}

//获取商城信息(通过storetype)
func (this *Engine) GetStoreInfo(args *rpcs.Args) *rpcs.Reply {
	res := GetStoreInfo(this.getArgs(args))
	return &rpcs.Reply{res, common.SC_OK}
}

//购买
func (this *Engine) Buy(args *rpcs.Args) *rpcs.Reply {
	res := Buy(this.getArgs(args))
	return &rpcs.Reply{res, common.SC_OK}
}

//获取成就信息(通过honorType)
func (this *Engine) GetHonorInfo(args *rpcs.Args) *rpcs.Reply {
	res := GetHonorInfo(this.getArgs(args))
	return &rpcs.Reply{res, common.SC_OK}
}

//成就领奖
func (this *Engine) HonorAward(args *rpcs.Args) *rpcs.Reply {
	res := HonorAward(this.getArgs(args))
	return &rpcs.Reply{res, common.SC_OK}
}

//获取创建房间所需元宝数
func (this *Engine) GetHYTWIngot(args *rpcs.Args) *rpcs.Reply {
	res := GetHYTWIngot(this.getArgs(args))
	return &rpcs.Reply{res, common.SC_OK}
}

//创建房间
func (this *Engine) CreateRoom(args *rpcs.Args) *rpcs.Reply {
	res := CreateRoom(this.getArgs(args))
	return &rpcs.Reply{res, common.SC_OK}
}

//坐下站起
func (this *Engine) SitDown(args *rpcs.Args) *rpcs.Reply {
	res := SitDown(this.getArgs(args))
	return &rpcs.Reply{res, common.SC_OK}
}

//是否革命
func (this *Engine) Revolution(args *rpcs.Args) *rpcs.Reply {
	res := Revolution(this.getArgs(args))
	return &rpcs.Reply{res, common.SC_OK}
}

//提示联邦出牌
func (this *Engine) HintFederal(args *rpcs.Args) *rpcs.Reply {
	res := HintFederal(this.getArgs(args))
	return &rpcs.Reply{res, common.SC_OK}
}

//断线重连获取比赛信息
func (this *Engine) Reconnect(args *rpcs.Args) *rpcs.Reply {
	res := Reconnect(this.getArgs(args))
	return &rpcs.Reply{res, common.SC_OK}
}

//获取道具效果
func (this *Engine) GetItemEffect(args *rpcs.Args) *rpcs.Reply {
	res := GetItemEffect(this.getArgs(args))
	return &rpcs.Reply{res, common.SC_OK}
}

//获取比赛中的道具
func (this *Engine) GetMatchItem(args *rpcs.Args) *rpcs.Reply {
	res := GetMatchItem(this.getArgs(args))
	return &rpcs.Reply{res, common.SC_OK}
}

//获取红点信息根据类型
func (this *Engine) GetRedPoint(args *rpcs.Args) *rpcs.Reply {
	res := GetRedPoint(this.getArgs(args))
	return &rpcs.Reply{res, common.SC_OK}
}

//获取邮件列表
func (this *Engine) GetMails(args *rpcs.Args) *rpcs.Reply {
	res := GetMails(this.getArgs(args))
	return &rpcs.Reply{res, common.SC_OK}
}

//读取邮件
func (this *Engine) ReadMail(args *rpcs.Args) *rpcs.Reply {
	res := ReadMail(this.getArgs(args))
	return &rpcs.Reply{res, common.SC_OK}
}

//领取邮件奖励
func (this *Engine) ReceMail(args *rpcs.Args) *rpcs.Reply {
	res := ReceMail(this.getArgs(args))
	return &rpcs.Reply{res, common.SC_OK}
}

//根据元宝获取匹配的房间类型
func (this *Engine) GetFMRoomType(args *rpcs.Args) *rpcs.Reply {
	res := GetFMRoomType(this.getArgs(args))
	return &rpcs.Reply{res, common.SC_OK}
}

//选一家联邦看牌
func (this *Engine) LookCard(args *rpcs.Args) *rpcs.Reply {
	res := LookCard(this.getArgs(args))
	return &rpcs.Reply{res, common.SC_OK}
}

//比赛中对其它玩家使用道具
func (this *Engine) MatchUseItem(args *rpcs.Args) *rpcs.Reply {
	res := MatchUseItem(this.getArgs(args))
	return &rpcs.Reply{res, common.SC_OK}
}

//获取正在比赛的房间ID
func (this *Engine) GetMatchingRoomID(args *rpcs.Args) *rpcs.Reply {
	res := GetMatchingRoomID(this.getArgs(args))
	return &rpcs.Reply{res, common.SC_OK}
}

//获取比赛规则
func (this *Engine) GetGameRule(args *rpcs.Args) *rpcs.Reply {
	res := GetGameRule(this.getArgs(args))
	return &rpcs.Reply{res, common.SC_OK}
}

//获取上班玩家的3信息
func (this *Engine) GetGTW(args *rpcs.Args) *rpcs.Reply {
	res := GetGTW(this.getArgs(args))
	return &rpcs.Reply{res, common.SC_OK}
}

//获取房间列表
func (this *Engine) GetRooms(args *rpcs.Args) *rpcs.Reply {
	res := GetRooms(this.getArgs(args))
	return &rpcs.Reply{res, common.SC_OK}
}

//进入房间
func (this *Engine) MatchingJudgment(args *rpcs.Args) *rpcs.Reply {
	res := MatchingJudgment(this.getArgs(args))
	return &rpcs.Reply{res, common.SC_OK}
}

//发牌
func (this *Engine) Deal(args *rpcs.Args) *rpcs.Reply {
	res := Deal(this.getArgs(args))
	return &rpcs.Reply{res, common.SC_OK}
}

//开牌
func (this *Engine) Begin(args *rpcs.Args) *rpcs.Reply {
	res := Begin(this.getArgs(args))
	return &rpcs.Reply{res, common.SC_OK}
}

//暂停
func (this *Engine) Pause(args *rpcs.Args) *rpcs.Reply {
	res := Pause(this.getArgs(args))
	return &rpcs.Reply{res, common.SC_OK}
}

//恢复
func (this *Engine) Resume(args *rpcs.Args) *rpcs.Reply {
	res := Resume(this.getArgs(args))
	return &rpcs.Reply{res, common.SC_OK}
}

//解散牌局
func (this *Engine) Dissolve(args *rpcs.Args) *rpcs.Reply {
	res := Dissolve(this.getArgs(args))
	return &rpcs.Reply{res, common.SC_OK}
}

//获取海选赛排名
func (this *Engine) GetAuditionInfo(args *rpcs.Args) *rpcs.Reply {
	res := GetAuditionInfo(this.getArgs(args))
	return &rpcs.Reply{res, common.SC_OK}
}

//获取实名认证信息
func (this *Engine) GetCertification(args *rpcs.Args) *rpcs.Reply {
	res := GetCertification(this.getArgs(args))
	return &rpcs.Reply{res, common.SC_OK}
}

//保存实名认证信息
func (this *Engine) SaveCertification(args *rpcs.Args) *rpcs.Reply {
	res := SaveCertification(this.getArgs(args))
	return &rpcs.Reply{res, common.SC_OK}
}

//获取海选赛奖励
func (this *Engine) GetAuditionAward(args *rpcs.Args) *rpcs.Reply {
	res := GetAuditionAward(this.getArgs(args))
	return &rpcs.Reply{res, common.SC_OK}
}

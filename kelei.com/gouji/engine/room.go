/*
房间
*/

package engine

import (
	"fmt"
	"strings"

	. "kelei.com/utils/common"
)

/*
游戏规则
默认版{
	1. 出牌时间15秒
	2. 自动出牌1次托管
}
录制版{
	1. 出牌时间30秒
	2. 自动出牌不托管
}
*/
const (
	GameRule_Normal = iota //默认版
	GameRule_Record        //录制版
)

const (
	Match_GJYX = iota //够级英雄
	Match_HYTW        //好友同玩
	Match_HXS         //海选赛
	Match_KDFS        //开点发四
)

const (
	RoomType_Primary      = iota //初级
	RoomType_Intermediate        //中级
	RoomType_Advanced            //高级
	RoomType_Master              //大师
	RoomType_Tribute             //进贡
)

const (
	SetController_NewCycle       = iota //新一轮
	SetController_Press                 //压牌
	SetController_Let                   //让牌
	SetController_Pass                  //要不了
	SetController_Burn                  //烧牌
	SetController_Burn_Press            //烧牌的状态下-出牌
	SetController_Burn_Fail             //烧牌失败
	SetController_ReverseBurn           //反烧
	SetController_NoChange              //没有变化
	SetController_LetBurn               //让牌烧牌
	SetController_Liuju                 //流局
	SetController_Kou            = 14   //扣牌
	SetController_Kou_ForceCheck = 15   //扣牌-要不起
	SetController_Burn_New       = 16   //烧牌新一轮出牌
)

const (
	Ranking_One   = iota //头科
	Ranking_Two          //二科
	Ranking_Three        //三科
	Ranking_Four         //四科
	Ranking_Five         //二落
	Ranking_Six          //大落
)

const (
	RoomStatus_Setout     = iota //准备
	RoomStatus_Deal              //发牌
	RoomStatus_Revolution        //革命
	RoomStatus_Liuju             //流局
	RoomStatus_Match             //开赛
)

const (
	MatchingStatus_Run   = iota //进行中
	MatchingStatus_Pause        //暂停
	MatchingStatus_Over         //结束
)

const (
	BurnStatus_Burn    = iota //正常烧牌
	BurnStatus_LetBurn        //让牌烧牌
)

const (
	PlayWaitTime      = 10 //要不起的等待时间
	PlayWaitTime_Long = 15 //其它的等待时间
)

type Room struct {
	id                string            //id
	matchid           int               //比赛类型
	roomtype          int               //房间类型
	pcount            int               //人数
	status            int               //房间状态
	matchingStatus    int               //开赛后的状态
	users             []*User           //玩家列表
	userids           []string          //玩家UserID集合
	idleusers         map[string]*User  //未落座玩家列表
	idleuserids       []string          //未落座玩家UserID集合
	cuser             *User             //牌权的玩家
	cards             []Card            //当前牌
	cardsuser         *User             //当前牌的玩家
	waitcusers        []*User           //等待牌权的玩家列表
	burnstatus        int               //烧牌状态
	ranking           []*User           //排名(头科、儿科、三科、四科、二拉、大拉)
	playTime          int               //出牌的次数
	playRound         int               //出牌的轮次
	users_cards       map[string]string //当前轮所有人的出牌信息
	inning            int               //当前局数
	innings           int               //总局数
	istribute         bool              //是否进贡
	isrevolution      bool              //是否可以革命
	setCtlMsg         []string          //设置牌权的内容,推送残局的时候用
	redIntegral       int               //红队好友同玩积分
	blueIntegral      int               //蓝队好友同玩积分
	surplusBKingCount int               //剩余大王数量
	surplusSKingCount int               //剩余小王数量
	surplusTwoCount   int               //剩余2数量
	multiple          int               //底分(倍数)
	playWaitTime      int               //要不起等待时间
	playWaitTime_Long int               //其它等待时间
	gameRule          int               //游戏规则
	firstController   *User             //第一个出牌的人
	gtwUserThreeInfo  string            //上班玩家的3信息
	judgmentUser      *User             //裁判
	records           []*string         //所有的记录（回放用）
	orderController   bool              //顺序牌权
	threeAskBelowUser bool              //打出够级牌,3轮够级牌后开始问下家要不要
	fake              bool              //假烧
	dealMode          int               //发牌模式
}

func (r *Room) GetRoomID() *string {
	return &r.id
}

func (r *Room) SetRoomID(roomid string) {
	r.id = roomid
}

//根据玩法规则配置房间
func (r *Room) configRoomByGameRule() {
	r.playWaitTime = PlayWaitTime
	r.playWaitTime_Long = PlayWaitTime_Long
	r.setGameRule(r.GetGameRuleConfig())
	if r.getGameRule() == GameRule_Record {
		r.playWaitTime = 10
		r.playWaitTime_Long = 30
		r.setOrderController(true)
		r.setThreeAskBelowUser(true)
		r.setFake(true)
	}
	r.setOrderController(true)
	r.setThreeAskBelowUser(true)
	r.setFake(true)
}

//重置
func (r *Room) reset() {
	r.userids = nil
	r.setPlayTime(0)
	r.setPlayRound(0)
	r.setSurplusBKingCount(4)
	r.setSurplusSKingCount(4)
	r.setSurplusTwoCount(16)
	r.setControllerUser(nil)
	r.setCurrentCards([]Card{})
	r.setCurrentCardsUser(nil)
	r.setWaitControllerUsers(nil)
	r.setSetCtlMsg([]string{})
	r.ranking = make([]*User, pcount)
	for _, user := range r.getUsers() {
		if user != nil {
			user.resume()
		}
	}
	r.users_cards = make(map[string]string, pcount)
}

//设置房间的基础信息
func (r *Room) setRoomBaseInfo() {
	allRoomData := *r.getAllRoomData()
	arrAllRoomData := strings.Split(allRoomData, "|")
	for _, roomData := range arrAllRoomData {
		arrRoomData_s := strings.Split(roomData, "$")
		arrRoomData := StrArrToIntArr(arrRoomData_s)
		roomType, _, multiple := arrRoomData[0], arrRoomData[1], arrRoomData[2]
		if roomType == r.GetRoomType() {
			r.setMultiple(multiple)
			break
		}
	}
}

//获取比赛是否已开启
func (r *Room) isMatching() bool {
	if r.GetRoomStatus() == RoomStatus_Setout {
		return false
	}
	return true
}

//获取游戏规则
func (r *Room) getGameRule() int {
	return r.gameRule
}

//设置游戏规则
func (r *Room) setGameRule(gameRule int) {
	r.gameRule = gameRule
}

//获取游戏规则
func (r *Room) getDealMode() int {
	return r.dealMode
}

//设置游戏规则
func (r *Room) setDealMode(dealMode int) {
	r.dealMode = dealMode
}

//获取游戏规则
func (r *Room) getMatchingStatus() int {
	return r.matchingStatus
}

//设置游戏规则
func (r *Room) setMatchingStatus(matchingStatus int) {
	r.matchingStatus = matchingStatus
}

//获取是否顺序牌权
func (r *Room) getOrderController() bool {
	return r.orderController
}

//设置是否顺序牌权
func (r *Room) setOrderController(orderController bool) {
	r.orderController = orderController
}

//获取是否3轮够级牌后开始问下家要不要
func (r *Room) getThreeAskBelowUser() bool {
	return r.threeAskBelowUser
}

//设置获取是否3轮够级牌后开始问下家要不要
func (r *Room) setThreeAskBelowUser(threeAskBelowUser bool) {
	r.threeAskBelowUser = threeAskBelowUser
}

//获取是否假烧
func (r *Room) getFake() bool {
	return r.fake
}

//设置是否假烧
func (r *Room) setFake(fake bool) {
	r.fake = fake
}

//获取裁判
func (r *Room) getJudgmentUser() *User {
	return r.judgmentUser
}

//设置裁判
func (r *Room) setJudgmentUser(judgmentUser *User) {
	r.judgmentUser = judgmentUser
}

//获取房间底分
func (r *Room) getMultiple() int {
	return r.multiple
}

//设置房间底分
func (r *Room) setMultiple(multiple int) {
	r.multiple = multiple
}

//更新出牌的轮次
func (r *Room) updatePlayRound() int {
	r.playRound += 1
	return r.playRound
}

//获取出牌的轮次
func (r *Room) getPlayRound() int {
	return r.playRound
}

//设置出牌的轮次
func (r *Room) setPlayRound(playRound int) {
	r.playRound = playRound
}

//更新出牌的次数
func (r *Room) updatePlayTime() int {
	r.playTime += 1
	return r.playTime
}

//获取出牌的次数
func (r *Room) getPlayTime() int {
	return r.playTime
}

//获取出牌的次数
func (r *Room) setPlayTime(playTime int) {
	r.playTime = playTime
}

//获取剩余大王的数量
func (r *Room) getSurplusBKingCount() int {
	return r.surplusBKingCount
}

//设置剩余大王的数量
func (r *Room) setSurplusBKingCount(v int) {
	r.surplusBKingCount = v
}

//更新剩余大王的数量
func (r *Room) updateSurplusBKingCount() {
	r.surplusBKingCount = r.surplusBKingCount - 1
}

//获取剩余小王的数量
func (r *Room) getSurplusSKingCount() int {
	return r.surplusSKingCount
}

//设置剩余小王的数量
func (r *Room) setSurplusSKingCount(v int) {
	r.surplusSKingCount = v
}

//更新剩余小王的数量
func (r *Room) updateSurplusSKingCount() {
	r.surplusSKingCount = r.surplusSKingCount - 1
}

//获取剩余2的数量
func (r *Room) getSurplusTwoCount() int {
	return r.surplusTwoCount
}

//设置剩余2的数量
func (r *Room) setSurplusTwoCount(v int) {
	r.surplusTwoCount = v
}

//更新剩余2的数量
func (r *Room) updateSurplusTwoCount() {
	r.surplusTwoCount = r.surplusTwoCount - 1
}

//获取设置牌权的命令
func (r *Room) getSetCtlMsg() []string {
	return r.setCtlMsg
}

//设置牌权的内容,推送残局时候用
func (r *Room) setSetCtlMsg(setCtlMsg []string) {
	r.setCtlMsg = setCtlMsg
}

//获取初始牌数量是否完整
func (r *Room) initCardCountIsIntegrity() bool {
	return cardCount == perCapitaCardCount
}

//获取房间人数
func (r *Room) GetPCount() int {
	return r.pcount
}

//更新房间人数
func (r *Room) updatePCount(v int) {
	r.pcount = r.pcount + v
}

//获取房间观战人数
func (r *Room) GetIdlePCount() int {
	return len(r.idleusers)
}

//根据index获取玩家
func (r *Room) getUserByIndex(index int) *User {
	return r.users[index]
}

//获取房间入座人数
func (r *Room) getUserCount() int {
	count := 0
	for _, user := range r.users {
		if user != nil {
			count += 1
		}
	}
	return count
}

//获取准备中的玩家数量
func (r *Room) getSetoutCount() int {
	count := 0
	for _, user := range r.users {
		if user != nil {
			if user.getStatus() == UserStatus_Setout {
				count += 1
			}
		}
	}
	return count
}

/*
获取玩家UserID字符串集合
in:是否刷新
*/
func (r *Room) getUserIDs(args ...bool) []string {
	if len(args) > 0 {
		if args[0] {
			r.userids = nil
		}
	}
	if r.userids == nil {
		r.userids = []string{}
		for _, user := range r.users {
			if user != nil {
				r.userids = append(r.userids, *user.userid)
			}
		}
	}
	return r.userids
}

/*
获取未落座玩家UserID字符串集合
in:是否刷新
*/
func (r *Room) getIdleUserIDs(args ...bool) []string {
	if len(args) > 0 {
		if args[0] {
			r.idleuserids = nil
		}
	}
	if r.idleuserids == nil {
		r.idleuserids = []string{}
		for _, user := range r.idleusers {
			if user != nil {
				r.idleuserids = append(r.idleuserids, *user.getUserID())
			}
		}
	}
	return r.idleuserids
}

/*
获取(UserID+IdleUserID)字符串集合
in:是否刷新
*/
func (r *Room) getAllUserIDs() []string {
	userids := r.getUserIDs(true)
	idleuserids := r.getIdleUserIDs(true)
	userids = InsertStringSlice(userids, idleuserids, len(userids))
	return userids
}

//获取比赛类型
func (r *Room) GetMatchID() int {
	return r.matchid
}

//设置比赛类型
func (r *Room) setMatchID(matchID int) {
	r.matchid = matchID
}

//获取是否进贡
func (r *Room) GetTribute() bool {
	return r.istribute
}

//获取是否进贡
func (r *Room) setTribute(tribute bool) {
	r.istribute = tribute
}

//获取是否革命
func (r *Room) GetRevolution() bool {
	return r.isrevolution
}

//设置是否革命
func (r *Room) setRevolution(revolution bool) {
	r.isrevolution = revolution
}

//获取总轮次
func (r *Room) getInnings() int {
	return r.innings
}

//设置当前轮次
func (r *Room) setInnings(innings int) {
	r.innings = innings
}

//获取当前轮次
func (r *Room) getInning() int {
	return r.inning
}

//设置当前轮次
func (r *Room) setInning(inning int) {
	r.inning = inning
}

//获取房间类型
func (r *Room) GetRoomType() int {
	return r.roomtype
}

//设置房间类型
func (r *Room) setRoomType(roomType int) {
	r.roomtype = roomType
}

//获取牌权玩家
func (r *Room) getControllerUser() *User {
	return r.cuser
}

//设置牌权玩家
func (r *Room) setControllerUser(user *User) {
	r.cuser = user
}

//获取当前牌
func (r *Room) getCurrentCards() []Card {
	return r.cards
}

//设置当前牌
func (r *Room) setCurrentCards(cards []Card) {
	r.cards = cards
}

//获取当前牌的玩家
func (r *Room) getCurrentCardsUser() *User {
	return r.cardsuser
}

//设置当前牌的玩家
func (r *Room) setCurrentCardsUser(user *User) {
	r.cardsuser = user
}

//获取等待牌权的玩家列表
func (r *Room) getWaitControllerUsers() []*User {
	return r.waitcusers
}

//设置等待牌权的玩家列表
func (r *Room) setWaitControllerUsers(users []*User) {
	r.waitcusers = users
	for _, user := range users {
		user.setStatus(UserStatus_WaitBurn)
	}
}

//更新等待牌权的玩家列表
func (r *Room) updateWaitControllerUsers(users []*User) {
	r.waitcusers = users
}

//删除等待牌权的玩家
func (r *Room) removeWaitControllerUser(user *User) {
	waitcusers := r.getWaitControllerUsers()
	for i, waitcuser := range waitcusers {
		if user.getUserID() == waitcuser.getUserID() {
			waitcusers = append(waitcusers[:i], waitcusers[i+1:]...)
			r.updateWaitControllerUsers(waitcusers)
			break
		}
	}
}

//获取房间状态
func (r *Room) GetRoomStatus() int {
	return r.status
}

//设置房间状态
func (r *Room) SetRoomStatus(status int) {
	r.status = status
}

//获取烧牌状态(让牌烧牌、普通烧牌)
func (r *Room) getBurnStatus() int {
	return r.burnstatus
}

//设置烧牌状态(让牌烧牌、普通烧牌)
func (r *Room) setBurnStatus(status int) {
	r.burnstatus = status
}

//获取落座的所有玩家
func (r *Room) getUsers() []*User {
	return r.users
}

//获取未落座的所有玩家
func (r *Room) getIdleUsers() map[string]*User {
	return r.idleusers
}

//获取玩家排名
func (r *Room) getRanking() []*User {
	return r.ranking
}

/*
把房间中所有玩家在负载均衡服务器上的信息都删除
重置玩家
*/
func (r *Room) deleteUsersInfo() {
	users := r.getUsers()
	for _, user := range users {
		if user != nil {
			user.deleteUserInfo()
		}
	}
}

/*
重置房间中所有的玩家
*/
func (r *Room) resetUsers() {
	users := r.getUsers()
	for _, user := range users {
		if user != nil {
			user.reset()
		}
	}
}

//关闭房间
func (r *Room) close() {
	RoomManage.removeRoom(r)
}

//给裁判提送信息
func (r *Room) pushJudgment(funcName string, message string) {
	if judgmentUser := r.getJudgmentUser(); judgmentUser != nil {
		judgmentUser.push(funcName, &message)
	}
}

//设置所有人托管状态
func (r *Room) SetAllUsersTrusteeshipStatus(status bool) {
	for _, user := range r.getUsers() {
		if user != nil {
			user.trusteeship = status
		}
	}
}

/*
所有选手端是否在线
*/
func (r *Room) AllUsersOnlinePush() {
	for _, user := range r.getUsers() {
		if user != nil {
			status := 0
			if user.getOnline() {
				status = 1
			}
			r.pushJudgment("Online_Push", fmt.Sprintf("%s|%d", *user.getUserID(), status))
		}
	}
}

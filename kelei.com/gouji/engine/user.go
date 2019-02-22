/*
玩家
*/

package engine

import (
	"bytes"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"

	. "kelei.com/utils/common"
	. "kelei.com/utils/delaymsg"
	"kelei.com/utils/frame"
	"kelei.com/utils/logger"
)

//玩家类型
const (
	TYPE_USER     = iota //玩家
	TYPE_JUDGMENT        //裁判
)

//玩家当前状态
const (
	UserStatus_NoSetout            = iota //未准备
	UserStatus_Setout                     //已准备
	UserStatus_NoPass                     //没过牌
	UserStatus_Let                        //让牌
	UserStatus_WaitBurn                   //等待烧牌
	UserStatus_Burn                       //烧牌
	UserStatus_BurnFail                   //烧牌失败
	UserStatus_Pass                       //过牌
	UserStatus_Go                         //走科
	UserStatus_Burn_Press                 //烧牌压牌
	UserStatus_Gouji                      //够级
	UserStatus_NoSitDown                  //未落座
	UserStatus_WaitRevolution             //等待自己革命
	UserStatus_WaitOtherRevolution        //等待别人革命
	UserStatus_Revolution                 //已革命
	UserStatus_WaitKou                    //等待扣牌
	UserStatus_Kou                        //扣牌
	UserStatus_Jieshao                    //解烧
)

const (
	TeamMark_Red  = iota //红队
	TeamMark_Blue        //蓝队
)

const (
	MatchResult_Win  = 1  //胜
	MatchResult_Flat = 0  //平
	MatchResult_Lose = -1 //负
)

const (
	TributeStatus_Wait = iota //等待还贡
	TributeStatus_Done        //已还贡
)

const (
	PlayType_Normal  = iota //正常出牌
	PlayType_EndGame        //重新进游戏，残局下，当前轮的出牌信息
)

const (
	EffectType_MatchItem = 2003 //比赛中使用的道具
)

//排名基点
var basePoint []int = []int{4, 2, 0, 0, -2, -4, 4}

type User struct {
	uid                *string       //平台userid
	userid             *string       //游戏userid
	conn               net.Conn      //链接
	GateRpc            string        //网关地址
	sessionKey         string        //sessionkey
	token              string        //access_token
	secret             string        //secret
	headUrl            string        //头像地址
	room               *Room         //房间ID
	status             int           //玩家状态
	cards              []Card        //牌列表
	index              int           //座位编号
	teamMark           int           //团队标示
	ranking            int           //排名(从0开始)
	matchResult        int           //比赛结果(1胜 0平 -1负)
	dm                 *DelayMessage //倒计时
	matchID            int           //当前所在的比赛类型
	basePoint          int           //比赛获得基点
	autoTimes          int           //倒计时结束自动操作的次数
	trusteeship        bool          //是否托管
	integral           int           //好友同玩积分,单轮结束清0
	tributeStatus      int           //贡状态(1等待还贡 2已还贡)
	jinLaTributeCards  []Card        //进落贡的牌
	huanLaTributeCards []Card        //还落贡的牌
	revolutionStatus   int           //革命状态(0没革命 1等待革命 2已革命)
	isAI               bool          //是否是玩家逃跑后,变成的AI
	online             bool          //玩家是否在线
	chatLastTime       time.Time     //发言的最后时间,用来做冷却
	levelRound         map[int]bool  //用来获得够级过多少轮
	burnUser           *User         //被自己烧的玩家
	effectTypes        map[int]int   //使用的道具效果列表
	lookedUsers        []*User       //看牌的对象列表
	ifLetCard          bool          //是否让过牌
	ifGoStayCard       bool          //是否是走科后留牌状态
	itemEffect         string        //道具效果
	userType           int           //玩家类型
	firstSetAfterBurn  bool          //烧牌后的第一套牌
}

//获取redis中的key
func (u *User) getKey() string {
	return "user:" + *u.getUserID()
}

//添加进落贡的牌
func (u *User) getJinLaTributeCards() []Card {
	return u.jinLaTributeCards
}

//添加进落贡的牌
func (u *User) setJinLaTributeCards(cards []Card) {
	u.jinLaTributeCards = cards
}

//是否正在比赛
func (u *User) isMatching() bool {
	if u == nil {
		return false
	}
	if u.getRoom() != nil && u.getRoom().isMatching() {
		return true
	}
	return false
}

//获取玩家是否有操作权限
func (u *User) getHandlePerm() bool {
	//不在比赛中,不是活动的,已过牌
	if !u.isMatching() || !u.isActive() || u.isPass() || u.isLet() {
		return false
	}
	return u.currCtlIsSelf()
}

//当前牌权是不是自己
func (u *User) currCtlIsSelf() bool {
	room := u.getRoom()
	//房间当前牌权的玩家
	if room.getControllerUser() == u {
		return true
	}
	//等待获取牌权的玩家(烧牌)
	waitUsers := room.getWaitControllerUsers()
	for _, user := range waitUsers {
		if user == u {
			return true
		}
	}
	return false
}

//玩家是否已走科
func (u *User) isGo() bool {
	return u.getStatus() == UserStatus_Go
}

//玩家是否是活动的
func (u *User) isActive() bool {
	if u.isGo() || u.isRevolution() {
		return false
	}
	return true
}

//玩家是否是头(走科或者革命后对家出了>3套牌,之后就不是头了)
func (u *User) isHead() bool {
	//玩家走科不是底牌
	if u.isGo() && !u.getIfGoStayCard() {
		return false
	}
	//玩家革命
	if u.isRevolution() {
		goUsers := u.getRoom().getGoUsers()
		//走科的人数>=1
		if len(goUsers) >= 1 {
			if len(goUsers) == 1 {
				user := goUsers[0]
				if user.getIfGoStayCard() {
					return true
				}
			}
			return false
		}
		//对家够级的次数>3
		if u.getSymmetryUser().getLevelRoundCount() > 3 {
			return false
		}
	}
	return true
}

//获取牌的ID列表
func (u *User) getCardsID() *string {
	buff := bytes.Buffer{}
	for _, card := range u.getCards() {
		buff.WriteString(fmt.Sprintf("%d$", card.ID))
	}
	cardsid := RemoveLastChar(buff)
	return cardsid
}

//获取对家
func (u *User) getSymmetryUser() *User {
	index := u.getIndex()
	users := u.getRoom().getUsers()
	return users[getMappingIndexs()[index]]
}

//是否是对家
func (u *User) isSymmetry(user *User) bool {
	return u.getSymmetryUser() == user
}

//是否过牌
func (u *User) isPass() bool {
	return u.getStatus() == UserStatus_Pass
}

//是否是让牌
func (u *User) isLet() bool {
	return u.getStatus() == UserStatus_Let
}

//往玩家手里添加牌
func (u *User) addCards(cards []Card) {
	userCards := CardList{}
	userCards = u.getCards()
	for _, card := range cards {
		userCards = append(userCards, card)
	}
	sort.Sort(userCards)
	u.setCards(userCards)
}

//删除玩家手里的牌
func (u *User) updateUserCards(play_indexs []int) {
	userCards := u.getCards()
	for i := len(play_indexs) - 1; i >= 0; i-- {
		index := play_indexs[i]
		userCards = append(userCards[:index], userCards[index+1:]...)
	}
	for k := 0; k < len(userCards); k++ {
		userCards[k].Index = k
	}
	u.setCards(userCards)
}

//检测是否走科
func (u *User) checkGo() {
	userCards := u.getCards()
	//玩家走科
	if len(userCards) <= 0 {
		room := u.getRoom()
		room.setRanking(u, 1)
		if room.isChaos() {
			//设置所有人都是没过牌状态
			room.setAllUserNoPass()
		}
	}
}

//记录被烧的信息
func (u *User) byBurn() {

}

//重置玩家
func (u *User) reset() {
	u.setMatchResult(MatchResult_Flat)
	u.status = UserStatus_NoSetout
	u.autoTimes = 0
	u.trusteeship = false
	u.levelRound = map[int]bool{}
	u.closeCountDown()
	u.lookedUsers = []*User{}
	u.setIfGoStayCard(false)
	u.resume()
}

//关闭玩家倒计时
func (u *User) closeCountDown() {
	u.close_countDown_playCard()
	u.close_countDown_setOut()
}

func (u *User) GetConn() net.Conn {
	return u.conn
}

func (u *User) SetConn(conn net.Conn) {
	u.conn = conn
}

func (u *User) GetGateRpc() string {
	return u.GateRpc
}

func (u *User) SetGateRpc(gateRpc string) {
	u.GateRpc = gateRpc
}

func (u *User) GetSessionKey() string {
	return u.sessionKey
}

func (u *User) SetSessionKey(sessionKey string) {
	u.sessionKey = sessionKey
}

func (u *User) GetToken() string {
	return u.token
}

func (u *User) SetToken(token string) {
	u.token = token
}

func (u *User) GetSecret() string {
	return u.secret
}

func (u *User) SetSecret(secret string) {
	u.secret = secret
}

func (u *User) GetHeadUrl() string {
	return u.headUrl
}

func (u *User) SetHeadUrl(headUrl string) {
	u.headUrl = headUrl
}

func (u *User) getItemEffect() string {
	return u.itemEffect
}

func (u *User) setItemEffect(itemEffect string) {
	u.itemEffect = itemEffect
}

func (u *User) getEffectTypes() map[int]int {
	return u.effectTypes
}

func (u *User) setEffectTypes(effectTypes map[int]int) {
	u.effectTypes = effectTypes
}

func (u *User) getUserType() int {
	return u.userType
}

func (u *User) setUserType(userType int) {
	u.userType = userType
}

func (u *User) getBasePoint() int {
	return u.basePoint
}

func (u *User) setBasePoint(basePoint int) {
	u.basePoint = basePoint
}

func (u *User) getLookedUsers() []*User {
	//	if *u.getUserID() == "67" {
	//		x := "69"
	//		u.lookedUser = UserManage.GetUser(&x)
	//	}
	return u.lookedUsers
}

func (u *User) addLookedUser(user *User) {
	u.lookedUsers = append(u.lookedUsers, user)
}

func (u *User) getTeamMark() int {
	return u.teamMark
}

func (u *User) setTeamMark(teamMake int) {
	u.teamMark = teamMake
}

func (u *User) getIsAI() bool {
	return u.isAI
}

func (u *User) setIsAI(isAI bool) {
	u.isAI = isAI
}

func (u *User) getOnline() bool {
	return u.online
}

func (u *User) setOnline(online bool) {
	u.online = online
}

func (u *User) getChatLastTime() time.Time {
	return u.chatLastTime
}

func (u *User) setChatLastTime() {
	u.chatLastTime = time.Now()
}

func (u *User) getLevelRoundCount() int {
	return len(u.levelRound)
}

func (u *User) setLevelRound(round int) {
	u.levelRound[round] = true
}

func (u *User) resetLevelRound() {
	u.levelRound = map[int]bool{}
}

func (u *User) getMatchID() int {
	return u.matchID
}

func (u *User) setMatchID(matchID int) {
	u.matchID = matchID
}

func (u *User) getBurnUser() *User {
	return u.burnUser
}

func (u *User) setBurnUser(user *User) {
	u.burnUser = user
}

func (u *User) getTributeStatus() int {
	return u.tributeStatus
}

func (u *User) setTributeStatus(tributeStatus int) {
	u.tributeStatus = tributeStatus
}

func (u *User) isRevolution() bool {
	return u.getStatus() == UserStatus_Revolution
}

func (u *User) getHYTWIntegral() int {
	return u.integral
}

func (u *User) setHYTWIntegral(integral int) {
	u.integral = integral
}

func (u *User) getUserID() *string {
	return u.userid
}

func (u *User) setUserID(userid *string) {
	u.userid = userid
}

func (u *User) getUID() *string {
	return u.uid
}

func (u *User) setUID(uid *string) {
	u.uid = uid
}

func (u *User) setRoom(room *Room) {
	u.room = room
}

func (u *User) getRoom() *Room {
	return u.room
}

func (u *User) getIndex() int {
	return u.index
}

func (u *User) setIndex(index int) {
	u.index = index
}

func (u *User) getIfGoStayCard() bool {
	return u.ifGoStayCard
}

func (u *User) setIfGoStayCard(b bool) {
	u.ifGoStayCard = b
}

func (u *User) getIfLetCard() bool {
	return u.ifLetCard
}

func (u *User) setIfLetCard(b bool) {
	u.ifLetCard = b
}

func (u *User) userGo() {
	u.setStatus(UserStatus_Go)
	users := u.getRoom().getUsers()
	for _, user := range users {
		user.setIfGoStayCard(false)
	}
	u.setIfGoStayCard(true)
}

func (u *User) getStatus() int {
	return u.status
}

func (u *User) setStatus(status int) {
	u.status = status
}

func (u *User) getCards() []Card {
	return u.cards
}

func (u *User) setCards(cards []Card) {
	u.cards = cards
}

func (u *User) getRanking() int {
	return u.ranking
}

func (u *User) setRanking(ranking int) {
	u.ranking = ranking
}

func (u *User) getMatchResult() int {
	return u.matchResult
}

func (u *User) setMatchResult(matchResult int) {
	u.matchResult = matchResult
}

func (u *User) getFirstSetAfterBurn() bool {
	return u.firstSetAfterBurn
}

func (u *User) setFirstSetAfterBurn(firstSetAfterBurn bool) {
	u.firstSetAfterBurn = firstSetAfterBurn
}

//获取托管状态
func (u *User) getTrusteeship() bool {
	return u.trusteeship
}

//是否能压过当前牌
func (u *User) canExcel() bool {
	args := []string{*u.getUserID()}
	indexs := Hint(args)
	if *indexs == "-2" {
		return false
	}
	return true
}

//是否是自己
func (u *User) isSelf(user *User) bool {
	return *u.getUserID() == *user.getUserID()
}

//是否有头
func (u *User) haveHead() bool {
	symmetryUser := u.getSymmetryUser()
	if symmetryUser.isHead() {
		return true
	}
	return false
}

//给一个玩家推送残局
func (u *User) pushEndGame() {
	logger.Debugf("推送残局")
	room := u.getRoom()
	//推送房间的状态信息
	room.matchingPush(u)
	time.Sleep(time.Millisecond * 5)
	//推送比赛的信息
	u.pushMatchInfo()
	//玩家回来,设置为在线
	u.setOnline(true)
}

//推送比赛的信息
func (u *User) pushMatchInfo() {
	room := u.getRoom()
	//推送房间的革命信息
	u.revolutionPush()
	time.Sleep(time.Millisecond * 5)
	//推送房间的走科信息
	room.goPush(u)
	time.Sleep(time.Millisecond * 5)
	//推送此人剩余的牌
	u.pushSurplusCards()
	time.Sleep(time.Millisecond * 5)
	//推送当前轮的出牌信息
	u.pushCyclePlayCardInfo()
	time.Sleep(time.Millisecond * 5)
	//推送托管状态
	u.TG_Push()
	//推送当前出牌状态
	ctlMsg := room.getSetCtlMsg()
	if len(ctlMsg) > 0 {
		u.setController(ctlMsg[0])
	}
	//<=10张,联邦看牌
	u.pushLookCard()
	//推送记牌器信息
	u.pushRememberCard()
	//所有人剩余的牌数
	u.pushSurplusCardCount()
	//推送暂停状态
	u.pushPauseStatus()
	//革命推送
	if u.getStatus() == UserStatus_WaitRevolution {
		message := fmt.Sprintf(",%s,%d,%d", *u.getUserID(), UserStatus_WaitRevolution, 10)
		u.setController(message)
	}
	//裁判端
	if room.getGameRule() == GameRule_Record {
		room.pushJudgment("Online_Push", fmt.Sprintf("%s|%d", *u.getUserID(), 1))
	}
}

/*
	推送房间的革命信息
	push:[Revolution_Push] userid|1
*/
func (u *User) revolutionPush() {
	users := u.getRoom().getUsers()
	str := ""
	for _, user := range users {
		if user != nil {
			if user.getStatus() == UserStatus_Revolution {
				str = fmt.Sprintf("%s|1", *user.getUserID())
				break
			}
		}
	}
	u.push("Revolution_Push", &str)
}

/*
推送暂停状态
*/
func (u *User) pushPauseStatus() {
	room := u.getRoom()
	matchingStatus := strconv.Itoa(room.getMatchingStatus())
	u.push("Pause_Push", &matchingStatus)
}

/*
推送等待革命状态
*/
func (u *User) pushWaitRevolutionStatus() {
	res := "1"
	u.push("Begin_Push", &res)
}

/*
SCC_Push(所有人剩余的牌数)
push:数量|数量|数量|数量|数量|数量
*/
func (u *User) pushSurplusCardCount() {
	arr := []string{}
	users := u.getRoom().getUsers()
	for _, user := range users {
		arr = append(arr, strconv.Itoa(len(user.getCards())))
	}
	message := strings.Join(arr, "|")
	u.push("SCC_Push", &message)
}

//<=10张,联邦看牌
func (u *User) pushLookCard() {
	users := u.getFederal()
	for _, user := range users {
		user.lookCard()
	}
}

//此玩家走科（最后一套留的底牌）或革命后（前三套），对家够级过来
func (u *User) noHeadHandleGouji() {
	room := u.getRoom()
	//收到对家够级后直接检测自己下手的两个人是否 “烧牌(有头)” or “压牌(没头)”
	if u.checkBelow() {
		return
	} else {
		//牌面玩家新一轮出牌
		room.currentCardsUserNewCycle()
	}
}

/*
RmbCard_Push(推送记牌器信息)
des:1.没有任何的记牌器道具,不推送	2.数量=-1代表没有对应的记牌器道具
out:大王剩余数量|小王剩余数量|2剩余数量
*/
func (u *User) pushRememberCard() {
	room := u.getRoom()
	BKingCount := -1
	SKingCount := -1
	TwoCount := -1
	//有王的记忆
	if u.getKingMemory() != -1 {
		BKingCount = room.getSurplusBKingCount()
		SKingCount = room.getSurplusSKingCount()
	}
	//有2的记忆
	if u.getTwoMemory() != -1 {
		TwoCount = room.getSurplusTwoCount()
	}
	//如果没有任何的记牌器道具
	if BKingCount == -1 && SKingCount == -1 && TwoCount == -1 {
		return
	}
	message := fmt.Sprintf("%d|%d|%d", BKingCount, SKingCount, TwoCount)
	u.push("RmbCard_Push", &message)
}

//断线重连获取比赛信息
func (u *User) Reconnect() *string {
	res := ""
	room := u.getRoom()
	if room == nil {
		return &res
	}
	u.setOnline(true)
	if room.isMatching() {
		//推送比赛的信息
		u.pushMatchInfo()
	} else {
		//推送房间的状态信息
		room.matchingPush(u)
	}
	return &res
}

//推送此人剩余的牌
func (u *User) pushSurplusCards() {
	buffer := bytes.Buffer{}
	for _, card := range u.cards {
		buffer.WriteString(strconv.Itoa(card.ID))
		buffer.WriteString("|")
	}
	str := buffer.String()
	if str != "" {
		message := str[0 : len(str)-1]
		//给此人推送剩余的牌局
		u.push("Opening_Push", &message)
	}
}

//推送当前轮的出牌信息
func (u *User) pushCyclePlayCardInfo() {
	room := u.getRoom()
	users_cards := room.users_cards
	for userid, cardids := range users_cards {
		message := fmt.Sprintf("%s,%s,,,%d", userid, cardids, PlayType_EndGame)
		time.Sleep(time.Millisecond * 5)
		u.push("Play", &message)
	}
}

//暂停倒计时
func (u *User) pause() {
	u.dm.PauseTask()
}

//恢复倒计时
func (u *User) resume() {
	u.dm.ResumeTask()
}

//革命倒计时
func (u *User) countDown_revolution() {
	u.dm.AddTask(time.Now().Add(time.Second*10), "revolutionCountDown", func(args ...interface{}) {
		//不革命
		u.revolution(false)
	}, nil)
}

//关闭革命倒计时
func (u *User) close_countDown_revolution() {
	u.dm.RemoveTask("revolutionCountDown")
}

//准备倒计时
func (u *User) countDown_setOut(t time.Duration) {
	if u == nil {
		return
	}
	room := u.getRoom()
	//如果是录像就没有准备倒计时,直接准备
	if room.getGameRule() == GameRule_Record {
		u.setStatus(UserStatus_Setout)
		return
	}
	u.dm.AddTask(time.Now().Add(t), "setOutCountDown", func(args ...interface{}) {
		//网关切换成访问游戏服务器
		u.push("ExitMatch", &Res_Succeed)
		//推送玩家退出了房间
		u.push("ExitRoom_Push", &Res_Succeed)
		//退出房间
		ExitMatch([]string{*u.getUserID()})
	}, nil)
}

//关闭准备倒计时
func (u *User) close_countDown_setOut() {
	u.dm.RemoveTask("setOutCountDown")
}

//开启出牌倒计时
func (u *User) countDown_playCard(waitTime int) {
	if u.getTrusteeship() {
		waitTime = 1
	}
	t, _ := time.ParseDuration(strconv.Itoa(waitTime) + "s")
	u.close_countDown_playCard()
	logger.Debugf("开启出牌倒计时:%d", waitTime)
	u.dm.AddTask(time.Now().Add(t), "playCardCountDown", func(args ...interface{}) {
		u.timeEnd(waitTime)
	}, []interface{}{1, 2, 3})
}

//关闭出牌倒计时
func (u *User) close_countDown_playCard() {
	if u != nil {
		u.dm.RemoveTask("playCardCountDown")
	}
}

//出牌倒计时结束,自动操作
func (u *User) timeEnd(waitTime int) {
	room := u.getRoom()
	current_cards := room.getCurrentCards()
	args := []string{*u.getUserID()}
	//托管
	if u.getTrusteeship() {
		//压牌
		if len(current_cards) > 0 {
			if u.getStatus() == UserStatus_WaitBurn {
				if u.canBurn() {
					Burn(args)
				} else {
					NoBurn(args)
				}
			} else {
				u.trusteeshipPlayCard()
			}
		} else {
			//出牌
			u.trusteeshipPlayCard()
		}
	} else {
		//压牌
		if len(current_cards) > 0 {
			if u.getStatus() == UserStatus_WaitBurn { //等待烧牌,不烧
				NoBurn(args)
			} else if u.getStatus() == UserStatus_Burn_Press { //点击烧牌之后,没有出牌
				logger.Debugf("点击烧牌之后,没有出牌")
				if u.getRoom().getGameRule() == GameRule_Record {
					u.BurnFail()
				}
			} else {
				CheckCard(args)
			}
		} else {
			//出牌
			u.trusteeshipPlayCard()
		}
		if waitTime >= room.playWaitTime_Long {
			u.trusteeshipHandle()
		}
	}
}

//托管出牌
func (u *User) trusteeshipPlayCard() {
	args := []string{*u.getUserID()}
	indexs := Hint(args)
	if *indexs == "-2" {
		CheckCard(args)
		return
	}
	args = append(args, *indexs)
	PlayCard(args)
}

//托管处理
func (u *User) trusteeshipHandle() {
	u.autoTimes += 1
	//倒计时结束自动操作的次数>=1次,进行托管
	if u.autoTimes >= 1 {
		u.setTrusteeship(true)
	}
}

//设置玩家托管
func (u *User) setTrusteeship(status bool) {
	if u.getRoom().getGameRule() == GameRule_Record {
		return
	}
	if !u.isMatching() {
		return
	}
	if u.trusteeship == status {
		u.TG_Push()
		return
	}
	u.trusteeship = status
	if !u.trusteeship {
		u.close_countDown_playCard()
	}
	u.autoTimes = 0
	u.TG_Push()
}

/*
托管推送
out:托管状态(0不托1托)
*/
func (u *User) TG_Push() {
	status := u.trusteeship
	//托管推送
	funcName := "TG_Push"
	message := "0"
	//托管
	if status {
		message = "1"
		room := u.getRoom()
		//如果玩家是（当前控牌人 或 等待烧牌状态）,关闭倒计时,玩家立即出牌
		if *room.getControllerUser().getUserID() == *u.getUserID() || u.getStatus() == UserStatus_WaitBurn {
			u.close_countDown_playCard()
			u.timeEnd(0)
		}
	}
	u.push(funcName, &message)
}

func (u *User) setController(message string) {
	pushMessageToUsers("SetController_Push", []string{message}, []string{*u.getUserID()})
}

func (u *User) setControllerUsers(message string) {
	pushMessageToUsers("SetController_Push", []string{message}, u.getRoom().getUserIDs())
	u.getRoom().pushJudgment("SetController_Push", message)
}

func (u *User) setCtlUsers(userid string, userstatus string, setControllerStatus int) {
	message := fmt.Sprintf("%s|%s|,,%d,", userid, userstatus, setControllerStatus)
	u.setControllerUsers(message)
}

//给此玩家推送信息
func (u *User) push(funcName string, message *string) {
	go func() {
		defer func() {
			if p := recover(); p != nil {
				errInfo := fmt.Sprintf("push : { %v }", p)
				logger.Errorf(errInfo)
			}
		}()
		time.Sleep(time.Millisecond)
		conn := u.GetConn()
		xServer := frame.GetRpcxServer()
		msg := fmt.Sprintf("%s&%s&%s", *u.getUserID(), funcName, *message)
		logger.Debugf("推送数据:%s", msg)
		err := xServer.SendMessage(conn, "service_path", "service_method", nil, []byte(msg))
		logger.CheckError(err, fmt.Sprintf("failed to send messsage to %s : ", conn.RemoteAddr().String()))
	}()
}

//玩家关闭连接(从比赛中)
func (u *User) close() {
	room := u.getRoom()
	if room == nil {
		return
	}
	//没开赛的时候,退出房间
	if !room.isMatching() {
		if u.getStatus() == UserStatus_NoSetout || u.getStatus() == UserStatus_Setout {
			u.exitRoom()
			UserManage.RemoveUser(u)
		}
	} else {
		//开赛后,玩家托管,并设置玩家离线
		u.setOnline(false)
		room.checkMatchPause(u)
	}
}

//玩家关闭连接(从长连接中)
func (u *User) closeFromKeepAliveUsers() {
	UserManage.RemoveUserFromKeepAliveUsers(u)
}

//玩家进入房间
func (u *User) enterRoom(room *Room) {
	u.room = room
	room.idleusers[*u.getUserID()] = u
}

//玩家“手动找位置”坐下
func (u *User) sitDown(seatIndex int) {
	room := u.getRoom()
	//从未落座的玩家中删除
	delete(room.getIdleUsers(), *u.getUserID())
	//落座
	if room.users[seatIndex] == nil {
		u.setIndex(seatIndex)
		room.users[seatIndex] = u
		room.updatePCount(1)
		u.setStatus(UserStatus_NoSetout)
	}
}

//玩家“自动找位置”坐下
func (u *User) sitDownAuto() {
	room := u.getRoom()
	for i, user := range room.getUsers() {
		if user == nil {
			u.sitDown(i)
			break
		} else {
			//人已经在房间中
			if *user.getUserID() == *u.getUserID() {
				break
			}
		}
	}
}

//玩家对号入座
func (u *User) sitDownPigeon(seatIndex int) {
	room := u.getRoom()
	for i, user := range room.getUsers() {
		if i == seatIndex {
			if user == nil {
				u.sitDown(i)
				break
			} else {
				//人已经在房间中
				if *user.getUserID() == *u.getUserID() {
					break
				}
			}
		}
	}
}

//玩家站起
func (u *User) standUp() {
	if u.getStatus() == UserStatus_NoSitDown {
		return
	}
	room := u.getRoom()
	//从已落座的玩家中删除
	room.users[u.getIndex()] = nil
	//放入未落座区域
	room.getIdleUsers()[*u.getUserID()] = u
	room.updatePCount(-1)
	u.setStatus(UserStatus_NoSitDown)
}

//玩家退出房间
func (u *User) exitRoom() {
	room := u.getRoom()
	if room == nil {
		return
	}
	if u.getStatus() == UserStatus_NoSitDown {
		delete(room.idleusers, *u.getUserID())
	} else {
		room.users[u.getIndex()] = nil
		room.updatePCount(-1)
	}
	totalPCount := room.GetIdlePCount() + room.GetPCount()
	if room.getGameRule() != GameRule_Record && totalPCount <= 0 {
		room.close()
	} else {
		u.exitRoomHandle()
	}
	u.room = nil
	u.reset()
}

//玩家退出房间的处理
func (u *User) exitRoomHandle() {
	room := u.getRoom()
	//好友同玩有人退出房间,房间重置
	if room.GetMatchID() == Match_HYTW {
		room.resetHYTW()
	}
	//房间的状态信息推送
	if room.GetMatchID() == Match_HXS {
		room.matchingHXSPush(nil)
	} else {
		room.matchingPush(nil)
	}
}

//退出正在比赛的房间
func (u *User) exitMatchRoom() {
	room := u.getRoom()
	//不能删！！！
	if room == nil {
		return
	}
	if room.getGameRule() == GameRule_Record {
		return
	}
	//增加逃跑次数
	u.updateFlee(1)
	//获取逃跑扣除的元宝和积分
	fleeCost := *FleeCost([]string{*u.getUserID()})
	arr := strings.Split(fleeCost, "|")
	ingot, _ := strconv.Atoi(arr[0])
	integral, _ := strconv.Atoi(arr[1])
	//扣元宝
	u.updateIngot(-ingot, 5)
	//扣积分
	u.updateIntegral(-integral)
	//删除玩家数据库中的记录信息
	u.deleteUserInfo()
	//将玩家改为AI,释放此玩家
	u.selfToAI()
}

//将玩家改为AI,释放此玩家
func (u *User) selfToAI() {
	u.setIsAI(true)
	u.setTrusteeship(true)
	u.updateUserID()
	UserManage.AddTheUser(u)
	fmt.Println(*u.getUserID(), "逃跑")
}

/*
UpdateUserID_Push
push:需要替换的UserID|新的UserID
*/
func (u *User) updateUserID() {
	room := u.getRoom()
	aiUserID := fmt.Sprintf("%s%s", *room.GetRoomID(), *u.getUserID())
	pushMessageToUsers("UpdateUserID_Push", []string{fmt.Sprintf("%s|%s", *u.getUserID(), aiUserID)}, room.getUserIDs())
	u.setUserID(&aiUserID)
	room.getUserIDs(true)
}

//获取联邦
func (u *User) getFederal() []*User {
	users := []*User{}
	roomUsers := u.getRoom().getUsers()
	for _, user := range roomUsers {
		//是联邦
		if user.getTeamMark() == u.getTeamMark() && *user.getUserID() != *u.getUserID() {
			users = append(users, user)
		}
	}
	return users
}

//获取联邦的userid
func (u *User) getFederalUserID() []string {
	userids := []string{}
	users := u.getFederal()
	for _, user := range users {
		userids = append(userids, *user.getUserID())
	}
	return userids
}

//获取活动的联邦
func (u *User) getActiveFederal() []string {
	userids := []string{}
	roomUsers := u.getFederal()
	for _, user := range roomUsers {
		userids = append(userids, *user.getUserID())
	}
	return userids
}

//获取我方的userid
func (u *User) getOurUserID() []string {
	userids := []string{}
	roomUsers := u.getRoom().getUsers()
	for _, user := range roomUsers {
		//是联邦
		if user.getTeamMark() == u.getTeamMark() {
			userids = append(userids, *user.getUserID())
		}
	}
	return userids
}

//提示联邦出牌
func (u *User) HintFederalPlayCard(federalUserID string, cardsIndex string) {
	pushMessageToUsers("HintFederal_Push", []string{fmt.Sprintf("%s,%s,%s", federalUserID, cardsIndex, *u.getUserID())}, []string{*u.getUserID(), federalUserID})
	u.getRoom().pushJudgment("HintFederal_Push", fmt.Sprintf("%s,%s,%s", federalUserID, cardsIndex, *u.getUserID()))
}

//选一家联邦看牌
func (u *User) LookCard(lookeduser *User) *string {
	res := Res_Unknown
	//走科或革命
	if !u.isActive() {
		//不是联邦
		if u.getTeamMark() != lookeduser.getTeamMark() {
			return &res
		}
		lookedUserCount := len(u.getLookedUsers())
		//看牌的人数已满
		if lookedUserCount >= 2 {
			return &res
		}
		//看牌
		look := func() {
			u.addLookedUser(lookeduser)
			lookeduser.lookCard()
			res = Res_Succeed
		}
		//没选择
		if lookedUserCount == 0 {
			look()
		} else { //已选择了一个玩家
			//已选择的那个玩家是否符合，再看一个玩家的条件
			firstLoodedUser := u.getLookedUsers()[0]
			//走科、小于10张
			if firstLoodedUser.isGo() || len(firstLoodedUser.getCards()) <= 10 {
				look()
			}
		}
	}
	return &res
}

//比赛中对其它玩家使用道具
func (u *User) MatchUseItem(dstuserid string, itemid int) *string {
	user := UserManage.GetUser(&dstuserid)
	if user == nil {
		return &Res_Unknown
	}
	res := u.useItemByItemID(itemid, 1, user.getRealityUserID())
	res = fmt.Sprintf("%s,%d", strings.Split(res, ",")[0], itemid)
	pushMessageToUsers("MUI_Push", []string{fmt.Sprintf("%s,%s,%d", *u.getUserID(), dstuserid, itemid)}, u.getRoom().getUserIDs())
	return &res
}

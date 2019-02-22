/*
玩家-操作-出牌
*/

package engine

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	. "kelei.com/utils/common"
	"kelei.com/utils/logger"
)

/*
玩家出牌
in:牌index列表
out:-1没有选择牌 -2你的手中没有这套牌 -3牌值无效 -4牌数不符 -5牌太小 -103玩家没有操作权限
	1成功
*/
func PlayCard(args []string) *string {
	fmt.Print("")
	res := Res_Succeed
	userid := args[0]
	user := UserManage.GetUser(&userid)
	//玩家没有操作权限
	if !user.getHandlePerm() {
		return &Res_NoPerm
	}

	play_indexs_str := args[1]
	if play_indexs_str == "" {
		res = "-1"
		return &res
	}

	//将字符串的index数组转化成数字index数组
	play_indexs := StrArrToIntArr(strings.Split(play_indexs_str, "|"))
	sort.Ints(play_indexs)

	if user.checkIndex(play_indexs) {
		res = "-2"
		return &res
	}

	userCards := user.getCards()
	play_cards := indexsToCards(play_indexs, userCards)
	if !isValid(play_cards) {
		res = "-3"
		return &res
	}

	room := user.getRoom()
	current_cards := room.getCurrentCards()
	isBurnFail := false
	//压牌
	if len(current_cards) > 0 {
		if len(current_cards) != len(play_cards) {
			res = "-4"
			return &res
		}
		if !excel(current_cards, play_cards) {
			res = "-5"
			return &res
		}
	} else {
		//烧牌状态下,新的一轮出牌
		if user.getStatus() == UserStatus_Burn_Press {
			/*
				新的一轮出牌，也就是烧别人的第一套牌已经打过了
				之后的每一套牌都需要挂王（除了最后一套）
			*/
			//没挂王、不是最后一套
			if !haveKing(play_cards) && !user.isFinallyCard(play_cards) {
				user.BurnFail()
				isBurnFail = true
			}
		}
	}
	//关闭玩家操作倒计时
	user.close_countDown_playCard()
	//更新玩家手中的牌
	user.updateUserCards(play_indexs)
	//将牌打出
	user.play(play_cards, isBurnFail)
	//<=10张,队友看牌
	user.lookCard()
	//检测玩家走科
	user.checkGo()
	//检测比赛是否结束
	room.checkMatchingOver()
	//是烧牌失败
	if isBurnFail {
		return &res
	}
	//检测烧牌失败
	if user.checkBurnFail() {
		return &res
	}
	//将牌打出之后,后续的操作
	time.Sleep(time.Millisecond * 20)
	user.playAfter(play_cards)
	user.setFirstSetAfterBurn(false)
	return &res
}

/*
LookCard_Push(联邦看牌)
push:userid|cardid$cardid
des:1.革命或走科之后选一家看牌
	2.牌数小于10张
*/
func (u *User) lookCard() {
	messages := []string{}
	userids := []string{}
	users := u.getFederal()
	for _, user := range users {
		message := ""
		//联邦在看自己的牌
		if u.federalIfLookedSelfCard(user) {
			message = fmt.Sprintf("%s|%s", *u.getUserID(), *u.getCardsID())
		} else {
			//少于10张牌
			if u.isActive() && len(u.getCards()) <= 10 {
				message = fmt.Sprintf("%s|%s", *u.getUserID(), *u.getCardsID())
			}
		}
		if message != "" {
			messages = append(messages, message)
			userids = append(userids, *user.getUserID())
		}
	}
	if len(messages) > 0 {
		pushMessageToUsers("LookCard_Push", messages, userids)
	}
}

//联邦是否在看自己的牌
func (u *User) federalIfLookedSelfCard(user_ *User) bool {
	users := user_.getLookedUsers()
	for _, user := range users {
		if user == u {
			return true
		}
	}
	return false
}

//打出
func (u *User) play(playCards []Card, isBurnFail bool) {
	room := u.getRoom()
	userids := room.getUserIDs()
	cardids := *getCardIDs(playCards)
	lessThanTen := ""
	cardCount := len(u.getCards())
	if cardCount <= 10 {
		lessThanTen = strconv.Itoa(cardCount)
	}
	//处理解烧
	u.loosenBurn()
	//没有烧牌失败,更换当前牌面的玩家
	if !isBurnFail {
		room.setCurrentCards(playCards)
		room.setCurrentCardsUser(u)
	}
	levelRoundCount := ""
	//挂大王
	if u.cardInvincible(playCards) {
		levelRoundCount = "0"
	}
	//是级牌
	if u.isLevel(playCards) {
		u.setLevelRound(room.getPlayRound())
		//对家革命
		if u.symmetryIsRevolution() {
			if u.getLevelRoundCount() <= 3 {
				levelRoundCount = strconv.Itoa(u.getLevelRoundCount())
			}
		}
	}
	//记录玩家的出牌信息(只保存一轮),为了重新进游戏
	room.RecordUserPlayCardInfo(u, &cardids)
	//记牌器更新
	u.rememberCard(playCards)
	//记录出牌信息
	u.recordPlayCard(playCards)
	//向所有人推送出牌的信息
	message := fmt.Sprintf("%s,%s,%s,%s,%d", *u.getUserID(), cardids, lessThanTen, levelRoundCount, PlayType_Normal)
	pushMessageToUsers("Play", []string{message}, userids)
	room.pushJudgment("Play", message)
}

//处理解烧
func (u *User) loosenBurn() {
	room := u.getRoom()
	currentCardsUser := room.getCurrentCardsUser()
	//当前牌面玩家是烧牌的状态,并且是自己的对家
	if currentCardsUser != nil && currentCardsUser.getStatus() == UserStatus_Burn_Press && (u.isSymmetry(currentCardsUser) || !u.haveHead()) {
		//将烧牌的玩家状态从烧牌出牌的状态改为没过牌
		logger.Debugf("处理解烧")
		currentCardsUser.setStatus(UserStatus_NoPass)
		//解烧,获取被烧的玩家
		burnUser := currentCardsUser.getBurnUser()
		u.setCtlUsers(*burnUser.getUserID(), strconv.Itoa(UserStatus_Jieshao), SetController_NoChange)
		logger.Debugf("解烧:%s,%s,%s", *currentCardsUser.getUserID(), *burnUser.getUserID(), *u.getUserID())
	}
}

//记牌器更新
func (u *User) rememberCard(playCards []Card) {
	room := u.getRoom()
	for _, card := range playCards {
		if card.Priority < Priority_Two {
			break
		}
		if card.Priority == Priority_Two {
			room.updateSurplusTwoCount()
		} else if card.Priority == Priority_SKing {
			room.updateSurplusSKingCount()
		} else if card.Priority == Priority_BKing {
			room.updateSurplusBKingCount()
		}
	}
}

//记录出牌信息
func (u *User) recordPlayCard(play_cards []Card) {
	room := u.getRoom()
	//记录出牌的信息（debug）
	playTime := room.updatePlayTime()
	bf := bytes.Buffer{}
	for _, card := range play_cards {
		bf.WriteString(strconv.Itoa(card.Value))
		bf.WriteString("|")
	}
	bf1 := bytes.Buffer{}
	for _, card := range u.getCards() {
		bf1.WriteString(strconv.Itoa(card.Value))
		bf1.WriteString("|")
	}
	//记录牌
	logger.Debugf("第（%d）次出牌:%s,剩余牌:%s", playTime, bf.String(), bf1.String())
}

//打出之后
func (u *User) playAfter(playCards []Card) {
	room := u.getRoom()
	if room == nil {
		return
	}
	//初始话房间为默认值
	room.setBurnStatus(BurnStatus_Burn)
	//比赛结束
	if !u.isMatching() {
		return
	}
	//如果打出的牌无敌
	if u.cardInvincible(playCards) {
		room.currentCardsUserNewCycle()
		return
	}
	//获取哪个玩家获取牌权
	var nextUser *User
	if u.isLevel(playCards) { //是够级
		//找对家
		symmetryHandle := func() bool {
			nextUser = u.activateSymmetry()
			//对家是走科留的最后一套底牌
			if nextUser.getIfGoStayCard() {
				nextUser.noHeadHandleGouji()
				return true
			}
			return false
		}
		//检测下手的两个人是否可以 “烧牌(有头)” or “压牌(没头)” or "压牌(有头,牌面玩家已打出3套够级牌)"
		if u.checkBelow() {
			return
		} else {
			if symmetryHandle() {
				return
			}
		}
	} else { //不是够级
		//乱打
		if room.isChaos() {
			room.setAllUserNoPass()
		}
		//获取下一顺位的玩家
		nextUser = room.getNextUser()
		//一个人走科留下底牌后,除了他的上手其它人都过牌了，上手（此人）出牌，没有下一个玩家可接牌
		if nextUser == nil {
			//如果对家还活着，激活对家
			nextUser = u.activateSymmetry()
			//如果对家不是活着的
			if !nextUser.isActive() {
				//出牌人（此人）新一轮出牌
				room.currentCardsUserNewCycle()
				return
			}
		}
	}
	room.setController(nextUser, SetController_Press)
}

//牌是否无敌
func (u *User) cardInvincible(playCards []Card) bool {
	card := playCards[0]
	if card.Priority == Priority_BKing {
		return true
	}
	return false
}

//够级激活对家出牌
func (u *User) activateSymmetry() *User {
	symmetryUser := u.getSymmetryUser()
	//对家是活动的
	if symmetryUser.isActive() {
		symmetryUser.setStatus(UserStatus_NoPass)
	}
	return symmetryUser
}

//获取和对家中间没有头也没过牌的玩家
func (u *User) getNoHeadUser() *User {
	roomUsers := u.getRoom().getUsers()
	symmetryUser := u.getSymmetryUser()
	users := getBetweenUsers(roomUsers, u, symmetryUser)
	var user *User
	for _, user_ := range users {
		if !user_.haveHead() && user_.getStatus() == UserStatus_NoPass {
			user = user_
			break
		}
	}
	return user
}

//是不是玩家最后一套牌
func (u *User) isFinallyCard(cards []Card) bool {
	return len(cards) == len(u.getCards())
}

//构成够级的张数列表
var numberList = []int{5, 4, 3, 2, 2, 1, 1, 1}

//是否够级
func (u *User) isLevel(cards []Card) bool {
	room := u.getRoom()
	//是乱打了
	if room.isChaos() {
		return false
	}
	currentCardsUser := room.getCurrentCardsUser()
	if currentCardsUser == nil {
		return false
	}
	//玩家没有头
	if !currentCardsUser.haveHead() {
		return false
	}
	//自己走了
	if currentCardsUser.isGo() && !currentCardsUser.getIfGoStayCard() {
		return false
	}
	return u.isLevelCard(cards)
}

//是否是级牌
func (u *User) isLevelCard(cards []Card) bool {
	kingCount := 0
	baseCardPriority := cards[len(cards)-1].Priority
	baseCardNumber := 0
	for _, card := range cards {
		if card.Priority > Priority_Two { //王
			kingCount = kingCount + 1
		} else {
			baseCardNumber = baseCardNumber + 1
		}
	}
	if kingCount > 0 {
		return true
	}
	tenPriority := 8
	if baseCardPriority < tenPriority {
		return false
	}
	levelNumber := numberList[baseCardPriority-tenPriority]
	if baseCardNumber >= levelNumber {
		return true
	}
	return false
}

//对家是否革命
func (u *User) symmetryIsRevolution() bool {
	if u.getSymmetryUser().isRevolution() {
		return true
	}
	return false
}

//对家是否走科
func (u *User) symmetryIsGo() bool {
	if u.getSymmetryUser().isGo() {
		return true
	}
	return false
}

//将出的牌的index列表转化成对应的牌列表
func indexsToCards(indexs []int, userCards []Card) []Card {
	cards := make([]Card, len(indexs))
	for i, index := range indexs {
		cards[i] = userCards[index]
	}
	return cards
}

//选择的牌是否索引越界
func (u *User) checkIndex(indexs []int) bool {
	return indexs[len(indexs)-1]+1 > len(u.cards)
}

/*
检查牌值是否有效
1.(王和2)可以和(任意一套牌)组合
*/
func isValid(cards []Card) bool {
	//除了王和2以外的任意一套牌
	anyCard := -1
	for _, card := range cards {
		if card.Priority < Priority_Two {
			cardPriority := card.Priority
			if anyCard != -1 && anyCard != cardPriority {
				return false
			}
			anyCard = cardPriority
		}
	}
	return true
}

//检测牌是否能压过(当前的牌,要出的牌)
func excel(a, b []Card) bool {
	//如果要压过的牌面的最小牌小于2,所有的2当这个值用（挂2的出法）
	a_min := a[len(a)-1].Priority
	b_min := b[len(a)-1].Priority
	//是否将2转化成最小牌
	a_convert, b_convert := a_min < Priority_Two, b_min < Priority_Two
	for i := 0; i < len(a); i++ {
		c := a[i].Priority
		//挂2,将2都转化成最小牌
		if a_convert && c == Priority_Two {
			c = a_min
		}
		d := b[i].Priority
		//挂2,将2都转化成最小牌
		if b_convert && d == Priority_Two {
			d = b_min
		}
		if d <= c {
			return false
		}
	}
	return true
}

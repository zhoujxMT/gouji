/*
房间比赛中
*/

package engine

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"

	. "kelei.com/utils/common"
	"kelei.com/utils/logger"
)

//获取牌的ID列表
func getCardIDs(cards []Card) *string {
	cardids := bytes.Buffer{}
	for _, card := range cards {
		cardids.WriteString(strconv.Itoa(card.ID))
		cardids.WriteString("|")
	}
	cardids_s := cardids.String()
	cardids_s = cardids_s[:len(cardids_s)-1]
	return &cardids_s
}

/*
设置牌权（真实牌权、等待牌权）
users是玩家：
	说明只有一个人获得 “真实牌权” ,给所有玩家推送的信息都是一样的
users是玩家数组：
	说明是多个人获得 “等待牌权” ,给所有玩家推送的信息是不一样的,这种情况只有两种可能
	1.级牌出牌之后,多人都可以（烧牌）
	2.（级牌让牌）之后,多人都可以上牌
*/
func (r *Room) setController(users interface{}, status int) {
	isArray := false
	userids := []string{}
	userstatuss := []int{}
	//获取玩家的操作
	userHandle := r.getUserHandle()
	if user, ok := users.(*User); ok {
		//只有最后一个玩家点出牌,才会找不到下一个出牌人
		if user == nil {
			logger.Debugf("比赛结束")
			return
		}
		//牌权发生变化,设置牌权
		if status != SetController_NoChange && status != SetController_Burn_Fail {
			if user.isRevolution() || user.getIfGoStayCard() { //此玩家革命后或走科最后一套的底牌,对家够级,牌权过来
				user.noHeadHandleGouji()
				return
			}
			r.setControllerUser(user)
		}
		userids = append(userids, *user.getUserID())
		if status == SetController_NewCycle {
			//新一轮
			r.newCycle()
			//更新轮次
			r.updatePlayRound()
		} else if status == SetController_Press {
			//压牌的情况,需要判断可不可以让牌
			if user.canLetCard() {
				status = SetController_Let
			} else {
				if !r.isChaos() && user == r.getCurrentCardsUser().getSymmetryUser() {
					user.setStatus(UserStatus_WaitKou)
					if user.canExcel() {
						status = SetController_Kou
					} else {
						status = SetController_Kou_ForceCheck
					}
				}
			}
		}
		userstatuss = append(userstatuss, status)
	} else if users, ok := users.([]*User); ok {
		isArray = true
		//设置等待牌权的玩家列表
		r.setWaitControllerUsers(users)
		//是顺序牌权
		if r.getOrderController() {
			users = []*User{users[0]}
		}
		for _, user := range users {
			userids = append(userids, *user.getUserID())
			canBurn := false
			canExcel := false
			if user.haveHead() {
				//假烧模式
				if r.getFake() {
					canExcel = user.canExcel()
				} else {
					canBurn = user.canBurn()
				}
				if !canExcel && !canBurn {
					userstatuss = append(userstatuss, SetController_Pass)
				} else {
					if r.getBurnStatus() == BurnStatus_Burn {
						userstatuss = append(userstatuss, SetController_Burn)
					} else if r.getBurnStatus() == BurnStatus_LetBurn {
						userstatuss = append(userstatuss, SetController_LetBurn)
					}
				}
			} else {
				r.setControllerUser(user)
				userstatuss = append(userstatuss, SetController_Press)
			}
		}
	} else {
		logger.Fatalf("传入的users参数无效")
	}
	//等待时间
	waitTime := r.getWaitTime(status)
	messages := []string{}
	roomUserids := r.getUserIDs()
	if isArray {
		for _, userid := range roomUserids {
			index := IndexStringOf(userids, &userid)
			var userstatus int
			if index != -1 {
				//相关的玩家,显示特定的信息
				userstatus = userstatuss[index]
			} else {
				//不相关的玩家,显示本来应该显示的信息
				userstatus = status
			}
			message := fmt.Sprintf("%s,%s,%d,%d", userHandle, strings.Join(userids, "|"), userstatus, waitTime)
			messages = append(messages, message)
		}
	} else {
		userid := userids[0]
		userstatus := userstatuss[0]
		//牌权不发生变化,userid设为空
		if status == SetController_NoChange {
			userid = ""
		}
		message := fmt.Sprintf("%s,%s,%d,%d", userHandle, userid, userstatus, waitTime)
		messages = append(messages, message)
	}
	//给玩家添加操作倒计时
	for _, userid := range userids {
		if status != SetController_NoChange {
			UserManage.GetUser(&userid).countDown_playCard(waitTime)
		}
	}
	r.setSetCtlMsg(messages)
	pushMessageToUsers("SetController_Push", messages, roomUserids)
	r.pushJudgment("SetController_Push", messages[0])
}

//牌面玩家占牌,新一轮出牌
func (r *Room) currentCardsUserNewCycle() {
	currentCardsUser := r.getCurrentCardsUser()
	r.setController(currentCardsUser, SetController_NoChange)
	setController := func(user *User) {
		time.Sleep(time.Millisecond * 1500)
		//如果牌面玩家的状态为烧牌出牌的状态
		if user.getStatus() == UserStatus_Burn_Press {
			//新一轮
			r.newCycle()
			r.setControllerUser(nil)
			r.setController(user, SetController_Burn_New)
		} else {
			r.setControllerUser(nil)
			r.setController(user, SetController_NewCycle)
		}
	}
	if currentCardsUser.isActive() { //此玩家没走科,自己获得牌权
		//去掉够级的提示
		setController(currentCardsUser)
	} else { //此玩家走科了,下手获得牌权
		r.newCycle()
		//为了获取下一个出牌人
		r.setControllerUser(currentCardsUser)
		r.setCurrentCardsUser(currentCardsUser)
		nextUser := r.getNextUser()
		//去掉够级的提示
		setController(nextUser)
	}
}

//获取操作倒计时
func (r *Room) getWaitTime(status int) int {
	waitTime := r.playWaitTime_Long
	//要不了 || 烧牌 || 让牌烧牌
	//	if status == SetController_Pass || status == SetController_Burn || status == SetController_LetBurn {
	if status == SetController_Pass {
		waitTime = r.playWaitTime
	}
	return waitTime
}

//获取玩家操作的数据
func (r *Room) getUserHandle() string {
	userHandle := ""
	user := r.getControllerUser()
	currentCardsUser := r.getCurrentCardsUser()
	if user == nil {
		return userHandle
	}
	status := user.getStatus()
	//如果控制牌权的人的状态是
	if status == UserStatus_Kou || status == UserStatus_Let || status == UserStatus_Pass || status == UserStatus_Burn || status == UserStatus_BurnFail || status == UserStatus_Go {
		info := ""
		if status == UserStatus_Go {
			info = strconv.Itoa(user.getRanking())
		} else if status == UserStatus_Burn {
			//被自己烧的玩家
			info = *user.getBurnUser().getUserID()
		}
		userHandle = fmt.Sprintf("%s|%d|%s", *user.getUserID(), status, info)
	} else if status == UserStatus_NoPass || status == UserStatus_WaitBurn || status == UserStatus_Burn_Press || status == UserStatus_WaitKou { //压牌 || 等待扣牌
		//更改让牌的玩家为过牌
		letUser := r.updateLetToPass()
		if letUser != nil {
			userHandle = fmt.Sprintf("%s|%d|%s", *letUser.getUserID(), letUser.getStatus(), "")
		}
		currentCards := r.getCurrentCards()
		//提示够级
		if user.isLevel(currentCards) && user == currentCardsUser {
			if userHandle != "" {
				userHandle = userHandle + "-" + fmt.Sprintf("%s|%d|%s", *user.getUserID(), UserStatus_Gouji, "")
			} else {
				userHandle = fmt.Sprintf("%s|%d|%s", *user.getUserID(), UserStatus_Gouji, "")
			}
		}
	}
	return userHandle
}

//更改让牌的玩家为过牌
func (r *Room) updateLetToPass() *User {
	users := r.getUsers()
	for _, user := range users {
		if user.getStatus() == UserStatus_Let {
			user.setStatus(UserStatus_Pass)
			return user
		}
	}
	return nil
}

//获取下一顺位出牌人
func (r *Room) getNextUser() *User {
	controllerUser := r.getControllerUser()
	//烧牌发牌的时候,当前牌的
	currentCardsUser := r.getCurrentCardsUser()
	users := r.getUsers()
	index := controllerUser.getIndex()
	var user *User
	for i := 0; i < pcount-1; i++ {
		index = getNextUserIndex(index)
		u := users[index]
		//不是出牌人,是活动的
		if u.getUserID() != currentCardsUser.getUserID() && u.isActive() {
			//没过牌 || 烧牌压牌
			if u.getStatus() == UserStatus_NoPass || u.getStatus() == UserStatus_Burn_Press {
				user = u
				break
			} else if u.getStatus() == UserStatus_Let { //让牌
				u.setStatus(UserStatus_NoPass)
				user = u
				break
			} else if isSymmetry(u, currentCardsUser) { //对家
				if u.getStatus() == UserStatus_WaitKou { //等待扣牌
					//检测下手的两个人是否可以压牌
					if u.checkBelowIfPressCard() {
						continue
					}
					user = u
					break
				}
			}
		}
	}
	return user
}

//获取下一个玩家index
func getNextUserIndex(index int) int {
	index = index + 1
	if index > pcount-1 {
		index = 0
	}
	return index
}

//获取两个玩家中间的玩家
func getBetweenUsers(users []*User, user1 *User, user2 *User) []*User {
	betweenUsers := []*User{}
	index := user1.getIndex()
	endIndex := user2.getIndex()
	indexs := []string{}
	for {
		index = getNextUserIndex(index)
		if index == endIndex {
			break
		}
		indexs = append(indexs, strconv.Itoa(index))
	}
	for _, index_s := range indexs {
		index, _ := strconv.Atoi(index_s)
		for _, user := range users {
			if user.getIndex() == index {
				if user.isActive() {
					betweenUsers = append(betweenUsers, user)
				}
			}
		}
	}
	return betweenUsers
}

//获取两个玩家中间的未过牌的玩家
func getBetweenNopassUsers(users []*User, user1 *User, user2 *User) []*User {
	betweenUsers := getBetweenUsers(users, user1, user2)
	nopassUsers := []*User{}
	for _, user := range betweenUsers {
		if user.getStatus() == UserStatus_NoPass || user.getStatus() == UserStatus_Burn_Press {
			nopassUsers = append(nopassUsers, user)
		}
	}
	return nopassUsers
}

//6个人对家的映射关系
func getMappingIndexs() []int {
	var mappingIndexsn []int
	if pcount == 2 {
		mappingIndexsn = []int{1, 0}
	} else if pcount == 4 {
		mappingIndexsn = []int{2, 3, 0, 1}
	} else if pcount == 6 {
		mappingIndexsn = []int{3, 4, 5, 0, 1, 2}
	}
	return mappingIndexsn
}

//两个玩家是否是对家
func isSymmetry(user1 *User, user2 *User) bool {
	return user1.getSymmetryUser().getIndex() == user2.getIndex()
}

//根据优先级获取牌列表
func getCardsByPriority(cs []Card, priority int) []Card {
	cards := []Card{}
	for _, card := range cs {
		if card.Priority < priority {
			break
		}
		if card.Priority == priority {
			cards = append(cards, card)
		}
	}
	return cards
}

//根据优先级获取牌是否存在
func isExistByPriority(cards []Card, priority int) bool {
	exist := false
	for _, card := range cards {
		if card.Priority < priority {
			exist = false
			break
		}
		if card.Priority == priority {
			exist = true
			break
		}
	}
	return exist
}

//获取房间匹配信息
func (r *Room) getRoomMatchingInfo() ([]string, []string) {
	//已落座和未落座的玩家集合
	userids := r.getAllUserIDs()
	//获取已落座的玩家状态
	getUsersStatuss := func() string {
		bfStatuss := bytes.Buffer{}
		for _, user := range r.users {
			if user != nil {
				userInfo, err := redis.Ints(user.GetUserInfo("vip", "level"))
				logger.CheckError(err)
				vip, level := userInfo[0], userInfo[1]
				bfStatuss.WriteString(fmt.Sprintf("%s$%d$%d$%d$%d|", *user.userid, user.getStatus(), vip, level, user.getHYTWIntegral()))
			} else {
				bfStatuss.WriteString("|")
			}
		}
		strStatuss := bfStatuss.String()
		strStatuss = strStatuss[:len(strStatuss)-1]
		return strStatuss
	}
	//获取未落座的玩家状态
	getIdleUsersStatuss := func() string {
		bfStatuss := bytes.Buffer{}
		for _, user := range r.idleusers {
			bfStatuss.WriteString(fmt.Sprintf("%s|", *user.userid))
		}
		strStatuss := bfStatuss.String()
		if strStatuss != "" {
			strStatuss = strStatuss[:len(strStatuss)-1]
		}
		return strStatuss
	}
	statuss := []string{fmt.Sprintf("%s#%s#%d", getUsersStatuss(), getIdleUsersStatuss(), r.getInning())}
	return userids, statuss
}

//设置所有人都过牌
func (r *Room) setAllUserPass() {
	for _, user := range r.users {
		if user.isActive() {
			user.setStatus(UserStatus_Pass)
		}
	}
}

//设置所有人都没过牌
func (r *Room) setAllUserNoPass() {
	for _, user := range r.users {
		if user.isActive() {
			user.setStatus(UserStatus_NoPass)
		}
	}
}

//获取是头的玩家列表
func (r *Room) getMatchingUsers() []*User {
	users := []*User{}
	for _, user := range r.users {
		if user.isHead() {
			users = append(users, user)
		}
	}
	return users
}

//获取走科的玩家列表
func (r *Room) getGoUsers() []*User {
	users := []*User{}
	for _, user := range r.users {
		if user.isGo() {
			users = append(users, user)
		}
	}
	return users
}

//是否是烧牌新一轮
func (r *Room) isBurnnewCycle() bool {
	for _, user := range r.getUsers() {
		if user.getStatus() == UserStatus_Burn_Press {
			return true
		}
	}
	return false
}

//开启新的一轮
func (r *Room) newCycle() {
	r.setCurrentCards([]Card{})
	r.setCurrentCardsUser(nil)
	r.setWaitControllerUsers(nil)
	r.setBurnStatus(BurnStatus_Burn)
	for _, user := range r.users {
		if user.isActive() && user.getStatus() != UserStatus_Burn_Press {
			user.setStatus(UserStatus_NoPass)
			user.setIfLetCard(false)
		}
		//修改走科人的“走科留牌状态”
		if !r.isBurnnewCycle() {
			user.setIfGoStayCard(false)
		}
		//革命后对家出了3套牌
		if user.getLevelRoundCount() == 3 {
			user.setLevelRound(1000)
		}
	}
	//某一轮的出牌信息
	r.users_cards = make(map[string]string, pcount)
}

/*
	设置玩家排名
	rType:设置排名的类型[0:被走科 1:走科]
	被走科包括:[烧牌失败、被闷]
*/
func (r *Room) setRanking(user *User, rType int) {
	//已经有排名了
	if user.getRanking() != -1 {
		return
	}
	//设置玩家走科
	user.userGo()
	//获取是否有革命的人
	haveRevolutionUser := r.haveRevolutionUser()
	//按顺序排名
	for i := 0; i < pcount; i++ {
		index := i
		if rType == 0 {
			index = pcount - 1 - i
		}
		//有革命的人,跳过四科的坑,成为小落
		if haveRevolutionUser {
			if index == 3 {
				index = 4
			}
		}
		u := r.ranking[index]
		if u == nil {
			logger.Debugf("%s 走科:%d", *user.getUserID(), index)
			r.ranking[index] = user
			user.setRanking(index)
			break
		}
	}
}

/*
	设置玩家排名
	index:排名
*/
func (r *Room) setRankingByIndex(user *User, index int) {
	//已经有排名了
	if user.getRanking() != -1 {
		return
	}
	//设置玩家走科
	user.userGo()
	u := r.ranking[index]
	if u == nil {
		logger.Debugf("%s 走科:%d", *user.getUserID(), index)
		r.ranking[index] = user
		user.setRanking(index)
	}
}

//是否乱打了
func (r *Room) isChaos() bool {
	return len(r.getMatchingUsers()) <= chaosPCount
}

//记录玩家的出牌信息(只保存一轮),为了重新进游戏
func (r *Room) RecordUserPlayCardInfo(user *User, cardInfo *string) {
	r.users_cards[*user.getUserID()] = *cardInfo
}

/*
玩家-操作-烧牌
*/

package engine

import (
	"sort"
	"strings"
	"time"

	. "kelei.com/utils/common"
)

//烧牌
func Burn(args []string) *string {
	res := Res_Succeed
	userid := args[0]
	user := UserManage.GetUser(&userid)
	//玩家没有操作权限
	if !user.getHandlePerm() {
		return &Res_NoPerm
	}
	user.setStatus(UserStatus_Burn)
	user.setFirstSetAfterBurn(true)
	user.handleBurn()
	return &res
}

//处理烧牌
func (u *User) handleBurn() {
	room := u.getRoom()
	u.close_countDown_playCard()
	controllerUser := room.getControllerUser()
	currentCardsUser := room.getCurrentCardsUser()
	//记录被自己烧的玩家
	u.setBurnUser(currentCardsUser)
	//当前牌的对家
	currentCardsSymmetry := currentCardsUser.getSymmetryUser()
	//是否是正常的烧牌
	isNormalBurn := room.getBurnStatus() == BurnStatus_Burn
	//更改牌权为点击要不起的玩家,之后改回（为了通知所有人谁过牌了）
	handleControllerUser := func(f func(), ctrUser *User) {
		room.setControllerUser(u)
		f()
		room.setControllerUser(ctrUser)
	}
	//获取等待牌权的玩家
	waitcusers := room.getWaitControllerUsers()
	//没有烧牌的
	if len(waitcusers) <= 0 {
		//设置出牌人获得牌权
		currentCardsUserGetController := func() {
			//出牌人是烧牌的状态
			if currentCardsUser.getStatus() == UserStatus_Burn_Press {
				//新一轮
				room.newCycle()
				handleControllerUser(func() {
					u.setStatus(UserStatus_Pass)
					room.setController(currentCardsUser, SetController_Burn_Press)
					u.setStatus(UserStatus_NoPass)
				}, currentCardsUser)
			} else {
				handleControllerUser(func() {
					room.setController(currentCardsUser, SetController_NewCycle)
				}, currentCardsUser)
			}
		}
		//设置出牌人对家获得牌权
		currentCardsSymmetryGetController := func() {
			handleControllerUser(func() {
				room.setController(currentCardsSymmetry, SetController_Press)
			}, currentCardsSymmetry)
		}
		//处理牌权
		handleController := func() {
			//“出牌人”和“对家”之间的人过牌,牌权给出牌人的对家
			if len(getBetweenUsers([]*User{u}, currentCardsUser, currentCardsSymmetry)) > 0 {
				if currentCardsSymmetry.isRevolution() {
					currentCardsSymmetry.noHeadHandleGouji()
				} else {
					currentCardsSymmetry.setStatus(UserStatus_NoPass)
					currentCardsSymmetryGetController()
				}
			} else {
				//出牌人的对家是等待扣牌的状态,获得牌权
				if currentCardsSymmetry.getStatus() == UserStatus_WaitKou {
					currentCardsSymmetryGetController()
				} else {
					currentCardsUserGetController()
				}
			}
		}
		if isNormalBurn { //正常烧牌
			handleController()
		} else { //让牌烧牌
			currentCardsSymmetry.setStatus(UserStatus_WaitKou)
			currentCardsSymmetryGetController()
		}
		return
	}
	first_waitcuser := waitcusers[0]
	//第一顺位的人是烧牌状态
	if first_waitcuser.getStatus() == UserStatus_Burn {
		byBurnUser := room.getCurrentCardsUser()
		//被烧的人记录信息
		byBurnUser = byBurnUser
		//更新等待烧牌的其它玩家状态为没过牌
		for _, user := range waitcusers {
			user.setStatus(UserStatus_NoPass)
		}
		//设置此人为烧牌的状态
		first_waitcuser.setStatus(UserStatus_Burn)
		//设置牌权
		room.setControllerUser(first_waitcuser)
		room.setController(first_waitcuser, SetController_Burn_Press)
		//设置此人为烧牌出牌的状态
		first_waitcuser.setStatus(UserStatus_Burn_Press)
		//清空等待烧牌的玩家
		room.setWaitControllerUsers(nil)
	} else {
		if u.getStatus() == UserStatus_Pass {
			handleControllerUser(func() {
				//是顺序牌权
				if room.getOrderController() {
					room.setController(waitcusers, SetController_LetBurn)
				} else {
					room.setController(controllerUser, SetController_NoChange)
				}
			}, controllerUser)
		}
	}
}

//玩家是否能烧牌
func (u *User) canBurn() bool {
	//没有头,不能烧牌
	if !u.haveHead() {
		return false
	}
	indexs_s := Hint([]string{*u.getUserID()})
	//没有能压过的牌
	if *indexs_s == "-2" {
		return false
	}
	//将用来压的牌从牌列表中取出
	userCards := u.getCards()
	cards := make([]Card, len(userCards))
	copy(cards, userCards)
	indexs := strings.Split(*indexs_s, "|")
	indexs_i := StrArrToIntArr(indexs)
	sort.Ints(indexs_i)
	for i := len(indexs_i) - 1; i >= 0; i-- {
		index := indexs_i[i]
		cards = append(cards[:index], cards[index+1:]...)
	}
	//剩下的牌能否每套都挂王
	if alwaysHangKing(cards) == false {
		return false
	}
	return true
}

//检测下手的两个人是否可以 “烧牌(有头)” or “压牌(没头)” or "压牌(有头,牌面玩家已打出3套够级牌)"
func (u *User) checkBelow() bool {
	room := u.getRoom()
	roomUsers := room.getUsers()
	//牌面玩家
	currentCardsUser := room.getCurrentCardsUser()
	//牌面玩家的对家
	currentCardsUserSymmetry := currentCardsUser.getSymmetryUser()
	//烧牌的玩家集合
	burnUsers := []*User{}
	//对门
	symmetryUser := u.getSymmetryUser()
	//获取有权利烧牌的玩家列表
	users := getBetweenUsers(roomUsers, u, symmetryUser)
	for _, user := range users {
		/*
				1
			2		6
			3		5
				4
			第一种情况
				1够级 	2没头接牌		2过牌
				2过牌后需要判断2的下手的出牌情况，遇到牌面玩家的对家，结束
			第二种情况
				1够级	23不要	4过牌	5没头接牌		5过牌
				5过牌后需要判断5的下手的出牌情况，遇到牌面玩家，结束
		*/
		if user == currentCardsUserSymmetry || user == currentCardsUser {
			break
		}
		//没有头
		if !user.haveHead() {
			if len(burnUsers) == 0 { //没有烧牌的人
				//没过牌
				if user.getStatus() == UserStatus_NoPass {
					//获得牌权
					room.setController(user, SetController_Press)
					return true
				}
				continue
			} else { //有烧牌的人
				continue
			}
		}
		//(没过牌、让牌、烧牌压牌)的人
		if user.getStatus() == UserStatus_NoPass || user.getStatus() == UserStatus_Let || user.getStatus() == UserStatus_Burn_Press {
			//压牌(有头,牌面玩家已打出3套够级牌)
			if room.getThreeAskBelowUser() && currentCardsUser.getLevelRoundCount() > 3 {
				burnUsers = append(burnUsers, user)
			} else {
				if user.canBurn() {
					burnUsers = append(burnUsers, user)
				} else {
					if len(getBetweenUsers([]*User{user}, u, symmetryUser)) > 0 {
						user.setStatus(UserStatus_Pass)
					}
				}
			}
		}
	}
	if len(burnUsers) <= 0 {
		return false
	}
	if room.getThreeAskBelowUser() && currentCardsUser.getLevelRoundCount() > 3 {
		//设置房间的状态为让牌烧牌(能看见玩家接牌)
		room.setBurnStatus(BurnStatus_LetBurn)
		room.setController(burnUsers, SetController_LetBurn)
	} else {
		//设置房间的状态为正常烧牌
		room.setBurnStatus(BurnStatus_Burn)
		room.setController(burnUsers, SetController_Burn)
	}
	return true
}

//检测下手的两个人是否可以压牌
func (u *User) checkBelowIfPressCard() bool {
	room := u.getRoom()
	roomUsers := room.getUsers()
	//对门
	symmetryUser := u.getSymmetryUser()
	//获取有权利烧牌的玩家列表
	users := getBetweenUsers(roomUsers, u, symmetryUser)
	for _, user := range users {
		if user.getStatus() == UserStatus_NoPass {
			return true
		}
	}
	return false
}

//检测是不是烧失败了
func (u *User) checkBurnFail() bool {
	//烧牌状态下,新的一轮出牌
	if u.getStatus() != UserStatus_Burn_Press {
		return false
	}
	cards := u.getCards()
	checkAlwaysHangKing := func() bool {
		if alwaysHangKing(cards) == false {
			u.BurnFail()
			return true
		}
		return false
	}
	//假烧
	if u.getRoom().getFake() {
		//不是点击烧牌后的第一套牌
		//		if !u.getFirstSetAfterBurn() {
		//			if checkAlwaysHangKing() {
		//				return true
		//			}
		//		}
	} else {
		if checkAlwaysHangKing() {
			return true
		}
	}
	return false
}

//是否可以一直挂王,除了最后一套
func alwaysHangKing(cards []Card) bool {
	kingCount, noHangCount := getKingAndNoHangCount(cards)
	//烧失败了
	if !alwaysHangKingSucceed(kingCount, noHangCount) {
		return false
	}
	return true
}

//获取王和不能挂的牌的数量差
func alwaysHangKingSucceed(kingCount, noHangCount int) bool {
	if (kingCount - noHangCount) < -1 {
		return false
	}
	return true
}

//获取王和不能挂的牌的数量
func getKingAndNoHangCount(cards []Card) (int, int) {
	//王的数量
	kingCount := 0
	//不能挂的牌集合
	noHangCards := make(map[int]bool)
	for _, card := range cards {
		if card.Priority > 13 { //王
			kingCount = kingCount + 1
		} else if card.Priority < 13 { //A 3-13
			noHangCards[card.Priority] = true
		}
	}
	return kingCount, len(noHangCards)
}

//烧牌失败
func (u *User) BurnFail() {
	room := u.getRoom()
	room.setRanking(u, 0)
	room.setController(u, SetController_Burn_Fail)
	users := room.getUsers()
	time.Sleep(time.Millisecond * 10)
	//将牌权移到下一顺位
	room.setAllUserNoPass()
	//获得牌权的玩家
	var user *User
	//被烧的玩家
	burnUser := u.getBurnUser()
	if !burnUser.isGo() { //被烧的玩家没走科
		user = burnUser
	} else {
		index := burnUser.getIndex()
		//牌权落都被烧人的下一顺位出牌人
		for i := 0; i < pcount-1; i++ {
			index = getNextUserIndex(index)
			if users[index].isActive() {
				user = users[index]
				break
			}
		}
	}
	time.Sleep(time.Second)
	room.setControllerUser(nil)
	room.setController(user, SetController_NewCycle)
}

//牌里面有没有挂王
func haveKing(cards []Card) bool {
	card := cards[0]
	if card.Priority > 13 { //王
		return true
	}
	return false
}

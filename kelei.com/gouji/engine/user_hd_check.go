/*
玩家-操作-过牌
*/

package engine

import (
	. "kelei.com/utils/common"
)

/*
过牌
in:
out:-1
*/
func CheckCard(args []string) *string {
	res := Res_Succeed
	userid := args[0]
	user := UserManage.GetUser(&userid)
	//玩家没有操作权限
	if !user.getHandlePerm() {
		return &Res_NoPerm
	}
	user.close_countDown_playCard()
	room := user.getRoom()
	currentCards := room.getCurrentCards()
	currentCardsUser := room.getCurrentCardsUser()
	//新一轮出牌,没有当前牌的玩家(正常情况下不可能出现)
	if currentCardsUser == nil {
		return &Res_Unknown
	}
	currentCardsSymmetryUser := currentCardsUser.getSymmetryUser()
	//当前状态
	currentUserStatus := user.getStatus()
	//设置玩家过牌
	user.setStatus(UserStatus_Pass)
	//获取哪个玩家获取牌权
	var nextUser *User
	if user.isLevel(currentCards) && user.isSelf(currentCardsSymmetryUser) { //是够级,是对家
		//够级后的对家过牌,检测下手的两个人是否可以 “烧牌(有头)” or “压牌(没头)”
		if user.checkBelow() {
			//等待扣牌
			user.setStatus(UserStatus_WaitKou)
			return &res
		}
	} else {
		if user.isLevel(currentCards) { //是级牌，不是对家，说明是没头的玩家
			/*
					1
				2		6
				3		5
					4
				1够级	23不要	4过牌	5没头接牌		5过牌
				5过牌后需要判断5的下手的出牌情况
			*/
			if user.checkBelow() {
				return &res
			} else {
				//尝试激活当前牌的对家
				currentCardsUser.activateSymmetry()
				//当前牌的对家还活着
				if currentCardsSymmetryUser.isActive() {
					nextUser = currentCardsSymmetryUser
				}
			}
		} else { //不是级牌
			//获取下一顺位的玩家
			nextUser = room.getNextUser()
			//不是牌面玩家的对家点过牌，牌权回到牌面玩家的对家
			if !user.isSelf(currentCardsSymmetryUser) && nextUser == nil {
				//不是乱缠 && 牌面玩家没走科 && 对家没走科
				if !room.isChaos() && (currentCardsUser.isActive() || currentCardsUser.getIfGoStayCard()) && currentCardsSymmetryUser.isActive() {
					currentCardsSymmetryUser.setStatus(UserStatus_WaitKou)
					nextUser = currentCardsSymmetryUser
				}
			}
		}
	}
	//更改牌权为点击扣牌的玩家,之后改回（为了通知所有人谁扣牌了）
	handleControllerUser := func(f func()) {
		if currentUserStatus == UserStatus_WaitKou {
			user.setStatus(UserStatus_Kou)
		}
		room.setControllerUser(user)
		f()
		room.setControllerUser(currentCardsUser)
	}
	//更换牌权
	if nextUser != nil {
		if nextUser.isSelf(currentCardsSymmetryUser) { //下一个玩家是牌面玩家的对家
			//			getBetweenNopassUsers(room.getUsers(),nextUser,currentCardsSymmetryUser)
			room.setControllerUser(nextUser)
			canLetCard := nextUser.canLetCard()
			room.setControllerUser(user)
			if canLetCard {
				room.setController(nextUser, SetController_Press)
			} else {
				room.setController(nextUser, SetController_Kou)
			}
		} else { //不可以让牌
			room.setController(nextUser, SetController_Press)
		}
		//是对家,对家没走科
		if user.isSelf(currentCardsSymmetryUser) && !currentCardsUser.isGo() {
			//不是四户乱缠
			if !room.isChaos() {
				user.setStatus(UserStatus_WaitKou)
			}
		}
	} else {
		//新一轮
		//出牌人走了,所有人都过牌了,牌权落到出牌人的下一顺位
		if currentCardsUser.isGo() {
			currentCardsUser.setIfGoStayCard(false)
			room.newCycle()
			room.setControllerUser(currentCardsUser)
			room.setCurrentCardsUser(currentCardsUser)
			//新一轮的出牌人
			nextUser = room.getNextUser()
			user.setStatus(UserStatus_Pass)
			room.setControllerUser(user)
			if currentUserStatus == UserStatus_WaitKou {
				user.setStatus(UserStatus_Kou)
			}
			room.setController(nextUser, SetController_NewCycle)
		} else {
			//如果牌面玩家的状态为烧牌出牌的状态
			if currentCardsUser.getStatus() == UserStatus_Burn_Press {
				//新一轮
				room.newCycle()
				handleControllerUser(func() {
					room.setController(currentCardsUser, SetController_Burn_Press)
				})
			} else {
				if currentUserStatus == UserStatus_WaitKou {
					user.setStatus(UserStatus_Kou)
				}
				room.setController(currentCardsUser, SetController_NewCycle)
			}
		}
	}
	return &res
}

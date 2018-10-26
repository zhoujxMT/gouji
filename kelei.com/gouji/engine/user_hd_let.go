/*
玩家-操作-让牌
*/

package engine

import (
	. "kelei.com/utils/common"
)

//让牌
func LetCard(args []string) *string {
	res := Res_Succeed
	userid := args[0]
	user := UserManage.GetUser(&userid)
	//玩家没有操作权限
	if !user.getHandlePerm() {
		return &Res_NoPerm
	}
	room := user.getRoom()
	user.close_countDown_playCard()
	users := room.getUsers()
	currentCards := room.getCurrentCards()
	symmetryUser := user.getSymmetryUser()
	user.setStatus(UserStatus_Let)
	user.setIfLetCard(true)
	//是够级
	if user.isLevel(currentCards) {
		//设置房间的状态为让牌烧牌
		room.setBurnStatus(BurnStatus_LetBurn)
		users := getBetweenNopassUsers(users, user, symmetryUser)
		room.setController(users, SetController_LetBurn)
	} else {
		//不是够级，直接让到下一顺位
		nextUser := room.getNextUser()
		room.setController(nextUser, SetController_Press)
	}
	return &res
}

//是否可以让牌
func (u *User) canLetCard() bool {
	room := u.getRoom()
	currentCardsUser := room.getCurrentCardsUser()
	controllerUser := room.getControllerUser()
	if room.isChaos() { //是乱打
		if !u.getIfLetCard() && room.getNextUser() != nil { //没让过牌
			return true
		}
	} else { //不是乱打
		//是对家、对家是活动的
		if isSymmetry(currentCardsUser, controllerUser) && (currentCardsUser.isActive() || currentCardsUser.getIfGoStayCard()) {
			//下面有可以接牌的玩家
			if room.getNextUser() != nil {
				return true
			}
		}
	}
	return false
}

/*
玩家开赛前
*/

package engine

import (
	"strconv"

	. "kelei.com/utils/common"
	"kelei.com/utils/logger"
)

/*
是否革命
in:是否革命(0不革 1革命)
out:-1牌不符合革命要求 -2已开赛不能革命
	1成功
*/
func Revolution(args []string) *string {
	res := ""
	userid := args[0]
	isRevolution, err := strconv.Atoi(args[1])
	logger.CheckFatal(err)
	user := UserManage.GetUser(&userid)
	if user == nil {
		res = Res_Unknown
		return &res
	}
	//关闭革命计时器
	user.close_countDown_revolution()
	res = user.revolution(isRevolution == 1)
	return &res
}

//革命
func (u *User) revolution(isRevolution bool) string {
	res := Res_Succeed
	if u.getRoom().GetRoomStatus() == RoomStatus_Match {
		res = "-2"
		return res
	}
	//牌不符合革命要求
	if u.getStatus() != UserStatus_WaitRevolution {
		res = "-1"
		return res
	}
	if isRevolution { //革命
		u.setStatus(UserStatus_Revolution)
		u.setCtlUsers(*u.getUserID(), strconv.Itoa(UserStatus_Revolution), SetController_NoChange)
		room := u.getRoom()
		//上班的状态下，玩家选择革命，牌权给下家
		if room.getFirstController() == u {
			index := u.getIndex()
			nextUserIndex := getNextUserIndex(index)
			room.firstController = room.getUsers()[nextUserIndex]
		}
	} else { //不革命
		u.setStatus(UserStatus_NoPass)
	}
	//	time.Sleep(time.Second)
	//检测比赛开始
	u.getRoom().check_start()
	return res
}

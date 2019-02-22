package engine

import (
	"bytes"
	"strconv"
	"strings"
	"time"

	. "kelei.com/utils/common"
)

const (
	DEAL_NORMAL = iota //默认没有好坏牌
	DEAL_RED           //红方牌好
	DEAL_BLUE          //蓝方牌好
)

/*
获取房间列表
*/
func GetRooms(args []string) *string {
	buff := bytes.Buffer{}
	rooms := RoomManage.GetRooms()
	for _, room := range rooms {
		buff.WriteString(*room.GetRoomID())
		buff.WriteString("|")
	}
	res := *RowsBufferToString(buff)
	return &res
}

/*
裁判端进入房间
in:roomid
out:1
push:房间状态、所有人剩余牌、当前轮的出牌信息、当前出牌状态、所有人剩余牌数量
*/
var lastTime = time.Now()

func MatchingJudgment(args []string) *string {
	userid := args[0]
	roomid := args[1]
	room := RoomManage.GetRoom(roomid)
	if room == nil {
		return &Res_Unknown
	}
	if time.Now().Sub(lastTime) < time.Second {
		return &Res_Succeed
	}
	lastTime = time.Now()
	user := UserManage.GetUser(&userid)
	user.setUserType(TYPE_JUDGMENT)
	user.setRoom(room)
	room.setJudgmentUser(user)
	room.matchingPush(nil)
	if room.isMatching() {
		//推送房间的走科信息
		time.Sleep(time.Millisecond * 10)
		//推送房间的革命信息
		user.revolutionPush()
		time.Sleep(time.Millisecond * 5)
		//推送房间的走科信息
		room.goPush(user)
		time.Sleep(time.Millisecond * 5)
		//推送所有人剩余的牌
		pushAllUserSurplusCards(room)
		time.Sleep(time.Millisecond * 10)
		//推送当前轮的出牌信息
		user.pushCyclePlayCardInfo()
		time.Sleep(time.Millisecond * 5)
		//推送当前出牌状态
		ctlMsg := room.getSetCtlMsg()
		if len(ctlMsg) > 0 {
			user.setController(ctlMsg[0])
		}
		//所有人剩余的牌数
		user.pushSurplusCardCount()
		//暂停状态
		user.pushPauseStatus()
		//所有选手端是否在线
		room.AllUsersOnlinePush()
	} else if room.GetRoomStatus() == RoomStatus_Revolution {
		user.pushWaitRevolutionStatus()
	}
	return &Res_Succeed
}

/*
推送所有人剩余牌
*/
func pushAllUserSurplusCards(r *Room) {
	users := r.getUsers()
	messages := []string{}
	for _, user := range users {
		buffer := bytes.Buffer{}
		for _, card := range user.cards {
			buffer.WriteString(strconv.Itoa(card.ID))
			buffer.WriteString("|")
		}
		str := buffer.String()
		if str != "" {
			str = str[0 : len(str)-1]
		}
		messages = append(messages, str)
	}
	//给裁判推送所有人的牌局
	r.pushJudgment("Opening_Judgment_Push", strings.Join(messages, "$"))
}

/*
发牌
in:roomid
out:1
*/
func Deal(args []string) *string {
	roomid := args[1]
	mode := DEAL_NORMAL
	if len(args) > 1 {
		mode, _ = strconv.Atoi(args[2])
	}
	room := RoomManage.GetRoom(roomid)
	if room == nil {
		return &Res_Unknown
	}
	if !room.userEnough() {
		return &Res_Unknown
	}
	room.reset()
	room.setDealMode(mode)
	room.deal()
	return &Res_Succeed
}

/*
开牌
in:roomid
out:1
*/
func Begin(args []string) *string {
	//	userid := args[0]
	roomid := args[1]
	room := RoomManage.GetRoom(roomid)
	if room == nil {
		return &Res_Unknown
	}
	if !room.userEnough() {
		return &Res_Unknown
	}
	go room.begin()
	return &Res_Succeed
}

/*
暂停
in:roomid
out:1
push:[Pause_Push] 1
*/
func Pause(args []string) *string {
	//	userid := args[0]
	roomid := args[1]
	room := RoomManage.GetRoom(roomid)
	if room == nil {
		return &Res_Unknown
	}
	if !room.userEnough() {
		return &Res_Unknown
	}
	room.pause()
	return &Res_Succeed
}

/*
恢复
in:roomid
out:1
push:[Resume_Push] 1
*/
func Resume(args []string) *string {
	//	userid := args[0]
	roomid := args[1]
	room := RoomManage.GetRoom(roomid)
	if room == nil {
		return &Res_Unknown
	}
	if !room.userEnough() {
		return &Res_Unknown
	}
	room.resume()
	return &Res_Succeed
}

/*
解散牌局
in:roomid
out:1
push:[Dissolve_Push] 1
*/
func Dissolve(args []string) *string {
	roomid := args[1]
	room := RoomManage.GetRoom(roomid)
	if room == nil {
		return &Res_Unknown
	}
	if !room.userEnough() {
		return &Res_Unknown
	}
	room.dissolve()
	return &Res_Succeed
}

package rpcs

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"kelei.com/utils/common"
	"kelei.com/utils/frame"
	"kelei.com/utils/logger"
	. "kelei.com/utils/rpcs"
)

var (
	gameBox *GameBox
)

func Init() {
	gameBox = &GameBox{}
	go func() {
		time.Sleep(time.Millisecond * 1)
		for {
			gameBox.loadGameAddr()
			gameBox.saveGameInfo()
			mapPCountTostrPcount()
			time.Sleep(time.Second * 10)
		}
	}()
}

type LoadbalancerS struct {
}

func (this *LoadbalancerS) GetGateAddr(ctx context.Context, args *Args, reply *Reply) error {
	gateAddrReply := this.getGateAddr()
	reply.RS = gateAddrReply.RS
	reply.SC = gateAddrReply.SC
	return nil
}

func (this *LoadbalancerS) getGateAddr() *Reply {
	xclient := frame.NewRpcxClient(frame.GetArgs().RpcxClient)
	defer xclient.Close()
	args := &Args{}
	args.V = make(map[string]interface{})
	reply := &Reply{}
	err := xclient.Call(context.Background(), "GetGateAddr", args, reply)
	if err != nil {
		logger.Errorf("failed to call: %s", err.Error())
		reply.SC = common.SC_GATEERR
		return reply
	}
	return reply
}

/*
用来记录玩家正在比赛的服务器地址和房间号
重新进入应用时候用的
*/
func (this *LoadbalancerS) InsertUserInfo(ctx context.Context, rpcArgs *Args, reply *Reply) error {
	db := frame.GetDB("loadbalancer")
	args := common.ExtractArgs(rpcArgs)
	userid := args[0]
	gameaddr := args[1]
	roomid := args[2]
	matchid := args[3]
	roomtype := args[4]
	_, err := db.Exec("insert into userinfo(userid,gameaddr,roomid,matchid,roomtype) values(?,?,?,?,?)", userid, gameaddr, roomid, matchid, roomtype)
	logger.CheckFatal(err, "InsertUserInfo")
	return nil
}

/*
删除玩家正在比赛的服务器地址和房间号
*/
func (this *LoadbalancerS) DeleteUserInfo(ctx context.Context, rpcArgs *Args, reply *Reply) error {
	db := frame.GetDB("loadbalancer")
	args := common.ExtractArgs(rpcArgs)
	userid := args[0]
	_, err := db.Exec("delete from userinfo where userid=?", userid)
	logger.CheckFatal(err, "DeleteUserInfo")
	return nil
}

//删除数据库中在此game服务器上的玩家
func (this *LoadbalancerS) RemoveUserMatchInfo(ctx context.Context, rpcArgs *Args, reply *Reply) error {
	args := common.ExtractArgs(rpcArgs)
	remote := args[0]
	db := frame.GetDB("loadbalancer")
	_, err := db.Exec("delete from userinfo where gameaddr=?", remote)
	logger.CheckFatal(err, "RemoveUserMatchInfo")
	return nil
}

//删除数据库中在此game服务器上的玩家
func (this *LoadbalancerS) RemoveRoomMatchInfo(ctx context.Context, rpcArgs *Args, reply *Reply) error {
	args := common.ExtractArgs(rpcArgs)
	remote := args[0]
	db := frame.GetDB("loadbalancer")
	_, err := db.Exec("delete from roominfo where gameaddr=?", remote)
	logger.CheckFatal(err, "RemoveRoomMatchInfo")
	return nil
}

/*
===================================================================
同步所有类型的比赛人数
===================================================================
*/
//map["gameAddr"]map[matchID]map[roomType]pcount
var mapPCount = make(map[string]map[int]map[int]int)
var strPCount = "-c"

func (this *LoadbalancerS) SyncPCount(ctx context.Context, rpcArgs *Args, reply *Reply) error {
	args := common.ExtractArgs(rpcArgs)
	addr := args[0]
	if mapPCount[addr] == nil {
		mapPCount[addr] = make(map[int]map[int]int)
	}
	infos_s := args[1]
	infos := strings.Split(infos_s, "|")
	matchID, roomType, pcount := 0, 0, 0
	for _, info_s := range infos {
		info_s := strings.Split(info_s, "$")
		info := common.StrArrToIntArr(info_s)
		matchID = info[0]
		roomType = info[1]
		pcount = info[2]
		if mapPCount[addr][matchID] == nil {
			mapPCount[addr][matchID] = make(map[int]int)
		}
		mapPCount[addr][matchID][roomType] = pcount
	}
	reply.RS = &strPCount
	return nil
}

func mapPCountTostrPcount() {
	if len(mapPCount) == 0 {
		return
	}
	m := make(map[int]map[int]int)
	for _, mapMatch := range mapPCount {
		for matchID, mapRoom := range mapMatch {
			if m[matchID] == nil {
				m[matchID] = make(map[int]int)
			}
			for roomType, pcount := range mapRoom {
				m[matchID][roomType] = m[matchID][roomType] + pcount
			}
		}
	}
	buff := bytes.Buffer{}
	for matchID, mapRoom := range m {
		for roomType, pcount := range mapRoom {
			buff.WriteString(fmt.Sprintf("%d$%d$%d|", matchID, roomType, pcount))
		}
	}
	str := buff.String()
	str = str[:len(str)-1]
	strPCount = str
}

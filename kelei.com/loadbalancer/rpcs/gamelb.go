/*
游戏负载均衡
*/

package rpcs

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"kelei.com/utils/common"
	"kelei.com/utils/frame"
	"kelei.com/utils/logger"
	. "kelei.com/utils/rpcs"
)

const (
	pcount = 6
)

type GameNode struct {
	ip        string
	port      int
	matchid   int
	roomtype  int
	usercount int
}

type GameBox struct {
	gameNodes []GameNode
}

func (g *GameBox) addGameNode(gameNode GameNode) {
	g.gameNodes = append(g.gameNodes, gameNode)
}

/*
服务器地址是否存在
-1不存在
>=0存在
*/
func (g *GameBox) getGameAddrIndex(ip string, port int) int {
	index := -1
	for k, gameNode := range g.gameNodes {
		//服务器已存在
		if gameNode.ip == ip && gameNode.port == port {
			index = k
			break
		}
	}
	return index
}

//加载游戏服务器地址
func (g *GameBox) loadGameAddr() {
	db := frame.GetDB("loadbalancer")
	ip := ""
	port := 0
	status_u := []uint8{}
	rows, err := db.Query("select ip,port,status from gameinfo")
	logger.CheckFatal(err, "loadGameAddr")
	defer rows.Close()
	for rows.Next() {
		rows.Scan(&ip, &port, &status_u)
		status := int(status_u[0])
		//关闭服务器节点
		if status == 0 {
			//检测需要关闭的服务器,进行关闭
			for k := len(g.gameNodes) - 1; k >= 0; k-- {
				gameNode := g.gameNodes[k]
				//服务器地址对应的服务器节点存在
				if gameNode.ip == ip && gameNode.port == port {
					logger.Infof("关闭game服务器节点 ip:%s port:%d matchid:%d roomtype:%d", gameNode.ip, gameNode.port, gameNode.matchid, gameNode.roomtype)
					g.gameNodes = append(g.gameNodes[:k], g.gameNodes[k+1:]...)
				}
			}
		} else {
			//如果服务器地址不存在
			if g.getGameAddrIndex(ip, port) == -1 {
				//添加新服务器地址
				gameNode := GameNode{ip, port, 0, 0, 0}
				logger.Infof("添加game服务器地址:%s:%d", gameNode.ip, gameNode.port)
				g.addGameNode(gameNode)
			}
		}
	}
}

//向数据库同步服务器人数
func (g *GameBox) saveGameInfo() {
	db := frame.GetDB("loadbalancer")
	mapGameAddr := map[string]int{}
	for _, gameNode := range gameBox.gameNodes {
		gameAddr := fmt.Sprintf("%s:%d", gameNode.ip, gameNode.port)
		mapGameAddr[gameAddr] = mapGameAddr[gameAddr] + gameNode.usercount
	}
	for gameAddr, usercount := range mapGameAddr {
		arr := strings.Split(gameAddr, ":")
		ip := arr[0]
		port, _ := strconv.Atoi(arr[1])
		_, err := db.Exec("update gameinfo set usercount=? where ip=? and port=?", usercount, ip, port)
		logger.CheckFatal(err, "saveGameInfo")
	}
}

//同步游戏的数据
func (this *LoadbalancerS) SyncGameInfo(ctx context.Context, rpcArgs *Args, reply *Reply) error {
	usercount := "1"
	args := common.ExtractArgs(rpcArgs)
	for _, arg := range args {
		arr := strings.Split(arg, ":")
		ip := arr[0]
		arr_i := common.StrArrToIntArr(arr[1:])
		port, matchid, roomtype, usercount := arr_i[0], arr_i[1], arr_i[2], arr_i[3]
		isUpdate := false
		for k, gameNode := range gameBox.gameNodes {
			if gameNode.ip == ip && gameNode.port == port && gameNode.matchid == matchid && gameNode.roomtype == roomtype {
				isUpdate = true
				gameBox.gameNodes[k].usercount = usercount
			}
		}
		//不是更新,是新添加
		if isUpdate == false {
			gameNode := GameNode{ip, port, matchid, roomtype, 0}
			gameBox.addGameNode(gameNode)
		}
	}
	reply.RS = &usercount
	reply.SC = common.SC_OK
	return nil
}

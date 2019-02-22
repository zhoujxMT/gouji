/*
后台逻辑服务器
*/

package main

import (
	"flag"
	"fmt"

	"kelei.com/server/cmds"
	"kelei.com/server/push"
	"kelei.com/server/update"
	"kelei.com/utils/common"
	"kelei.com/utils/frame"
	"kelei.com/utils/logger"
	"kelei.com/utils/mysql"
)

var (
	serversAddr = flag.String("serversAddr", "tcp@localhost:9150", "server address")
	gameDB      = flag.String("gameDB", "game,root,111111,127.0.0.1:3306,gouji_game", "")
)

func main() {
	defer func() {
		if p := recover(); p != nil {
			err := fmt.Sprintf("server server crash:%v", p)
			common.SendMail("够级", err)
			logger.Fatalf(err)
		}
	}()
	//解析参数
	flag.Parse()
	//启动服务
	args := frame.Args{}
	args.ServerName = "server"
	args.Commands = cmds.GetCmds()
	//游戏服务器的客户端
	args.RpcxClient = &frame.RpcxClient{frame.Discovery_MultipleServers, frame.Rpcx{Addr: serversAddr}, "GoujiS", frame.UNIDIRECTIONAL}
	//mysql
	sqlDSNs := []*mysql.SqlDSN{}
	sqlDSNs = append(sqlDSNs, mysql.AnalysisFlag2SqlDSN(gameDB))
	args.Sql = &frame.Sql{sqlDSNs}
	//框架启动完毕后执行的方法
	args.Loaded = start
	//通过参数启动框架
	frame.Load(args)
}

func start() {
	push.Init()
	update.Init()
}

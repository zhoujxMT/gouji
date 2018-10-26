package main

import (
	"flag"
	"fmt"

	"kelei.com/gate/cmds"
	"kelei.com/gate/rpcs"
	"kelei.com/utils/common"
	"kelei.com/utils/frame"
	"kelei.com/utils/logger"
	"kelei.com/utils/mysql"
)

//var (
//	addr          = flag.String("addr", "localhost:10110", "server address")
//	foreignAddr   = flag.String("foreignAddr", "localhost:10110", "server address")
//	etcdAddr      = flag.String("etcdAddr", "localhost:2379", "etcd address")
//	basePath      = flag.String("base", "/rpcx", "prefix path")
//	wsAddr        = flag.String("wsAddr", "localhost:11201", "server address")
//	wsForeignAddr = flag.String("wsForeignAddr", "localhost:11201", "server address")
//	gameDB        = flag.String("gameDB", "game,root,111111,127.0.0.1:3306,richman_game", "")
//	memberDB      = flag.String("memberDB", "member,root,111111,127.0.0.1:3306,richman_member", "")
//)

var (
	addr          = flag.String("addr", "localhost:9250", "server address")
	foreignAddr   = flag.String("foreignAddr", "localhost:9250", "server address")
	etcdAddr      = flag.String("etcdAddr", "localhost:2379", "etcd address")
	basePath      = flag.String("base", "/rpcx", "prefix path")
	wsAddr        = flag.String("wsAddr", "localhost:9280", "server address")
	wsForeignAddr = flag.String("wsForeignAddr", "localhost:9280", "server address")
	certificate   = flag.String("certificate", "", "certificate")
	gameDB        = flag.String("gameDB", "game,root,111111,127.0.0.1:3306,gouji_game", "")
	memberDB      = flag.String("memberDB", "member,root,111111,127.0.0.1:3306,gouji_nmmember", "")
	lbDB          = flag.String("lbDB", "loadbalancer,root,111111,127.0.0.1:3306,gouji_loadbalancer", "")
)

func main() {
	defer func() {
		if p := recover(); p != nil {
			err := fmt.Sprintf("gate server crash:%v", p)
			common.SendMail("够级", err)
			logger.Fatalf(err)
		}
	}()
	//解析参数
	flag.Parse()
	//启动服务
	args := frame.Args{}
	args.ServerName = "gate"
	args.Commands = cmds.GetCmds()
	//loadbalancer的服务端
	args.RpcxServer = &frame.RpcxServer{frame.Discovery_Etcd, frame.Rpcx{Addr: addr, ForeignAddr: foreignAddr, EtcdAddr: etcdAddr, BasePath: basePath}, []interface{}{new(rpcs.GateS)}}
	//richman的客户端
	args.RpcxClient = &frame.RpcxClient{frame.Discovery_Etcd, frame.Rpcx{EtcdAddr: etcdAddr, BasePath: basePath}, "GoujiS", frame.BIDIRECTIONAL}
	//mysql
	sqlDSNs := []*mysql.SqlDSN{}
	sqlDSNs = append(sqlDSNs, mysql.AnalysisFlag2SqlDSN(gameDB))
	sqlDSNs = append(sqlDSNs, mysql.AnalysisFlag2SqlDSN(memberDB))
	sqlDSNs = append(sqlDSNs, mysql.AnalysisFlag2SqlDSN(lbDB))
	args.Sql = &frame.Sql{sqlDSNs}
	//websocket
	args.WebSocket = &frame.WebSocket{*wsAddr, *wsForeignAddr, *certificate, HandleClientData}
	args.Loaded = start
	frame.Load(args)
}

func start() {
}

/*
处理客户端数据
out:是否终止后续方法的调用
*/
func HandleClientData(funcName string, args []string) (res string, termination bool) {
	switch funcName {
	case "test":
	}
	return res, termination
}

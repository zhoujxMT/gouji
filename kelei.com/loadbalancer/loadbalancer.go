package main

import (
	"flag"
	"fmt"

	"kelei.com/loadbalancer/cmds"
	"kelei.com/loadbalancer/rpcs"
	"kelei.com/utils/common"
	"kelei.com/utils/frame"
	"kelei.com/utils/logger"
	"kelei.com/utils/mysql"
)

//var (
//	addr     = flag.String("addr", "localhost:10100", "server address")
//	etcdAddr = flag.String("etcdAddr", "localhost:2379", "etcd address")
//	basePath = flag.String("base", "/rpcx", "prefix path")
//)

var (
	addr     = flag.String("addr", "localhost:9350", "server address")
	etcdAddr = flag.String("etcdAddr", "localhost:2379", "etcd address")
	basePath = flag.String("base", "/rpcx", "prefix path")
	lbDB     = flag.String("lbDB", "loadbalancer,root,111111,127.0.0.1:3306,gouji_loadbalancer", "")
)

func main() {
	defer func() {
		if p := recover(); p != nil {
			err := fmt.Sprintf("loadbalancer server crash:%v", p)
			common.SendMail("够级", err)
			logger.Fatalf(err)
		}
	}()
	//解析参数
	flag.Parse()
	//启动服务
	args := frame.Args{}
	args.ServerName = "loadbalancer"
	args.Commands = cmds.GetCmds()
	//login的服务端
	args.RpcxServer = &frame.RpcxServer{frame.Discovery_Peer2Peer, frame.Rpcx{addr, nil, nil, nil}, []interface{}{new(rpcs.LoadbalancerS)}}
	//gate的客户端
	args.RpcxClient = &frame.RpcxClient{frame.Discovery_Etcd, frame.Rpcx{nil, nil, etcdAddr, basePath}, "GateS", frame.UNIDIRECTIONAL}
	//mysql
	sqlDSNs := []*mysql.SqlDSN{}
	sqlDSNs = append(sqlDSNs, mysql.AnalysisFlag2SqlDSN(lbDB))
	args.Sql = &frame.Sql{sqlDSNs}
	args.Loaded = start
	frame.Load(args)
}

func start() {
	rpcs.Init()
}

package main

import (
	"flag"
	"fmt"

	"kelei.com/login/cmds"
	. "kelei.com/login/modes"
	"kelei.com/utils/common"
	"kelei.com/utils/frame"
	"kelei.com/utils/logger"
	"kelei.com/utils/mysql"
)

var (
	addr        = flag.String("addr", "localhost:9480", "http address")
	etcdAddr    = flag.String("etcdAddr", "localhost:2379", "etcd address")
	basePath    = flag.String("base", "/rpcx", "prefix path")
	memberDB    = flag.String("memberDB", "member,root,111111,127.0.0.1:3306,gouji_nmmember", "member db")
	certificate = flag.String("certificate", "", "https certificate")
)

func main() {
	defer func() {
		if p := recover(); p != nil {
			err := fmt.Sprintf("gouji server crash:%v", p)
			common.SendMail("够级", err)
			logger.Fatalf(err)
		}
	}()
	//解析参数
	flag.Parse()
	//启动服务
	args := frame.Args{}
	args.ServerName = "login"
	args.Commands = cmds.GetCmds()
	//http服务
	args.HttpGin = &frame.HttpGin{Addr: *addr, Modes: loadModes(), Certificate: *certificate}
	//gate的客户端
	args.RpcxClient = &frame.RpcxClient{frame.Discovery_Etcd, frame.Rpcx{nil, nil, etcdAddr, basePath}, "GateS", frame.UNIDIRECTIONAL}
	//mysql
	sqlDSNs := []*mysql.SqlDSN{mysql.AnalysisFlag2SqlDSN(memberDB)}
	args.Sql = &frame.Sql{sqlDSNs}
	//通过参数启动框架
	frame.Load(args)
}

func loadModes() frame.Modes {
	modes := frame.Modes{}
	modes["Login"] = &Login{}
	return modes
}

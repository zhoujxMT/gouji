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

//var (
//	addr        = flag.String("addr", "localhost:10101", "http address")
//	lbAddr      = flag.String("lbAddr", "localhost:10100", "loadbalancer address")
//	memberDB    = flag.String("memberDB", "member,root,111111,127.0.0.1:3306,richman_member", "member db")
//	certificate = flag.String("certificate", "", "https certificate")
//)

var (
	addr        = flag.String("addr", "localhost:9480", "http address")
	lbAddr      = flag.String("lbAddr", "localhost:9350", "loadbalancer address")
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
	args.HttpGin = &frame.HttpGin{Addr: *addr, Modes: loadModes(), Certificate: *certificate}
	args.RpcxClient = &frame.RpcxClient{frame.Discovery_Peer2Peer, frame.Rpcx{lbAddr, nil, nil, nil}, "LoadbalancerS", frame.UNIDIRECTIONAL}
	sqlDSNs := []*mysql.SqlDSN{mysql.AnalysisFlag2SqlDSN(memberDB)}
	args.Sql = &frame.Sql{sqlDSNs}
	args.Loaded = start
	frame.Load(args)
}

func loadModes() frame.Modes {
	modes := frame.Modes{}
	modes["Login"] = &Login{}
	return modes
}

func start() {
}

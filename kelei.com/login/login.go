package main

import (
	"flag"

	"kelei.com/login/cmds"
	. "kelei.com/login/modes"
	"kelei.com/utils/frame"
)

var (
	addr   = flag.String("addr", "localhost:10101", "http address")
	lbAddr = flag.String("lbAddr", "localhost:10100", "server address")
)

func main() {
	//解析参数
	flag.Parse()
	//启动服务
	args := frame.Args{}
	args.ServerName = "login"
	args.Commands = cmds.GetCmds()
	args.HttpGin = &frame.HttpGin{*addr, loadModes()}
	args.RpcxClient = &frame.RpcxClient{frame.Discovery_Peer2Peer, frame.Rpcx{lbAddr, nil, nil, nil}, "LoadbalancerS", frame.UNIDIRECTIONAL}
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

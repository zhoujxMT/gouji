package main

import (
	"flag"

	"kelei.com/loadbalancer/cmds"
	"kelei.com/loadbalancer/rpcs"
	"kelei.com/utils/frame"
)

var (
	addr     = flag.String("addr", "localhost:10100", "server address")
	etcdAddr = flag.String("etcdAddr", "localhost:2379", "etcd address")
	basePath = flag.String("base", "/rpcx", "prefix path")
)

func main() {
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
	args.Loaded = start
	frame.Load(args)
}

func start() {
}

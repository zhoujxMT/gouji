package main

import (
	"flag"

	"kelei.com/gate/cmds"
	"kelei.com/gate/rpcs"
	"kelei.com/utils/frame"
	. "kelei.com/utils/rpcs"
)

var (
	addr          = flag.String("addr", "localhost:10110", "server address")
	foreignAddr   = flag.String("foreignAddr", "localhost:10110", "server address")
	etcdAddr      = flag.String("etcdAddr", "localhost:2379", "etcd address")
	basePath      = flag.String("base", "/rpcx", "prefix path")
	wsAddr        = flag.String("wsAddr", "localhost:11201", "server address")
	wsForeignAddr = flag.String("wsForeignAddr", "localhost:11201", "server address")
)

func main() {
	//解析参数
	flag.Parse()
	//启动服务
	args := frame.Args{}
	args.ServerName = "gate"
	args.Commands = cmds.GetCmds()
	//loadbalancer的服务端
	args.RpcxServer = &frame.RpcxServer{frame.Discovery_Etcd, frame.Rpcx{addr, foreignAddr, etcdAddr, basePath}, []interface{}{new(rpcs.GateS)}}
	//richman的客户端
	args.RpcxClient = &frame.RpcxClient{frame.Discovery_Etcd, frame.Rpcx{nil, nil, etcdAddr, basePath}, "RichManS", frame.BIDIRECTIONAL}
	//websocket
	args.WebSocket = &frame.WebSocket{*wsAddr, *wsForeignAddr, frame.WSType_ws, Init, HandleClientData}
	args.Loaded = start
	frame.Load(args)
}

func start() {
}

//初始化,获取userid
func Init(uid string) (userid string, err error) {
	userid = uid
	return userid, err
}

/*
处理客户端数据
out:是否终止后续方法的调用
*/
func HandleClientData(args *Args) (res string, termination bool) {
	termination = false
	return res, termination
}

package main

import (
	"flag"

	"kelei.com/richman/cmds"
	"kelei.com/richman/engine"
	"kelei.com/richman/rpcs"
	"kelei.com/utils/frame"
)

var (
	addr        = flag.String("addr", "localhost:10120", "server address")
	foreignAddr = flag.String("foreignAddr", "localhost:10120", "server address")
	etcdAddr    = flag.String("etcdAddr", "localhost:2379", "etcd address")
	basePath    = flag.String("base", "/rpcx", "prefix path")
)

func main() {
	//解析参数
	flag.Parse()
	//启动服务
	args := frame.Args{}
	args.ServerName = "richman"
	args.Commands = cmds.GetCmds()
	args.RpcxServer = &frame.RpcxServer{frame.Discovery_Etcd, frame.Rpcx{addr, foreignAddr, etcdAddr, basePath}, []interface{}{new(rpcs.RichManS)}}
	args.Loaded = start
	frame.Load(args)
}

func start() {
	engine := engine.NewEngine()
	inject(engine)
}

func inject(engine *engine.Engine) {
	rpcs.Inject(engine)
	cmds.Inject(engine)
}

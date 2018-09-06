package main

import (
	"flag"

	"kelei.com/template-normal/cmds"
	"kelei.com/utils/frame"
)

var (
	addr = flag.String("addr", "localhost:12000", "server address")
)

func main() {
	//解析参数
	flag.Parse()
	//启动服务
	args := frame.Args{}
	args.ServerName = "template"
	args.Commands = cmds.GetCmds()
	args.Loaded = start
	frame.Load(args)
}

func start() {
}

package main

import (
	"flag"
	"fmt"

	"kelei.com/gouji/cmds"
	eng "kelei.com/gouji/engine"
	"kelei.com/gouji/rpcs"
	"kelei.com/utils/common"
	"kelei.com/utils/frame"
	"kelei.com/utils/logger"
	"kelei.com/utils/mysql"
	"kelei.com/utils/redis"
)

var (
	addr        = flag.String("addr", "localhost:9150", "server address")
	foreignAddr = flag.String("foreignAddr", "localhost:9150", "server address")
	lbAddr      = flag.String("lbAddr", "localhost:9350", "loadbalancer address")
	etcdAddr    = flag.String("etcdAddr", "localhost:2379", "etcd address")
	basePath    = flag.String("base", "/rpcx", "prefix path")
	gameDB      = flag.String("gameDB", "game,root,111111,127.0.0.1:3306,gouji_game", "")
	memberDB    = flag.String("memberDB", "member,root,111111,127.0.0.1:3306,gouji_nmmember", "")
	gameRedis   = flag.String("gameRedis", "localhost:9100", "")
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
	args.ServerName = "gouji"
	args.Commands = cmds.GetCmds()
	args.RpcxServer = &frame.RpcxServer{frame.Discovery_Etcd, frame.Rpcx{addr, foreignAddr, etcdAddr, basePath}, []interface{}{new(rpcs.GoujiS)}}
	args.RpcxClient = &frame.RpcxClient{frame.Discovery_Peer2Peer, frame.Rpcx{lbAddr, nil, nil, nil}, "LoadbalancerS", frame.UNIDIRECTIONAL}
	//redis
	redisDSNs := []*redis.RedisDSN{}
	redisDSNs = append(redisDSNs, &redis.RedisDSN{"game", *gameRedis, ""})
	args.Redis = &frame.Redis{redisDSNs}
	//mysql
	sqlDSNs := []*mysql.SqlDSN{}
	sqlDSNs = append(sqlDSNs, mysql.AnalysisFlag2SqlDSN(gameDB))
	sqlDSNs = append(sqlDSNs, mysql.AnalysisFlag2SqlDSN(memberDB))
	args.Sql = &frame.Sql{sqlDSNs}
	args.Loaded = start
	frame.Load(args)
}

func start() {
	engine := eng.NewEngine(*addr)
	//将引擎注入到各模块
	inject(engine)
	//裁判端定制
	judgment()
}

func inject(engine *eng.Engine) {
	rpcs.Inject(engine)
	cmds.Inject(engine)
}

func judgment() {
	room := eng.Room{}
	gameRule := room.GetGameRuleConfig()
	if gameRule == eng.GameRule_Record {
		eng.RoomManage.Judgment()
	}
}

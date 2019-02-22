package main

import (
	"flag"
	"fmt"

	redis_ "github.com/garyburd/redigo/redis"

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
	etcdAddr    = flag.String("etcdAddr", "localhost:2379", "etcd address")
	basePath    = flag.String("base", "/rpcx", "prefix path")
	gameDB      = flag.String("gameDB", "game,root,111111,127.0.0.1:3306,gouji_game", "")
	memberDB    = flag.String("memberDB", "member,root,111111,127.0.0.1:3306,gouji_nmmember", "")
	gameRedis   = flag.String("gameRedis", "localhost:9100", "")
	memberRedis = flag.String("memberRedis", "localhost:9400", "")
)

func main() {
	defer func() {
		if p := recover(); p != nil {
			err := fmt.Sprintf("gouji server crash:%v", p)
			common.SendMail("够级", err)
			unBindRedis()
			logger.Fatalf(err)
		}
	}()
	//解析参数
	flag.Parse()
	//启动服务
	args := frame.Args{}
	args.ServerName = "gouji"
	args.Commands = cmds.GetCmds()
	//gate的服务端
	args.RpcxServer = &frame.RpcxServer{frame.Discovery_Etcd, frame.Rpcx{addr, foreignAddr, etcdAddr, basePath}, []interface{}{new(rpcs.GoujiS)}}
	//redis
	redisDSNs := []*redis.RedisDSN{}
	redisDSNs = append(redisDSNs, &redis.RedisDSN{"game", *gameRedis, ""})
	redisDSNs = append(redisDSNs, &redis.RedisDSN{"member", *memberRedis, ""})
	args.Redis = &frame.Redis{redisDSNs}
	//mysql
	sqlDSNs := []*mysql.SqlDSN{}
	sqlDSNs = append(sqlDSNs, mysql.AnalysisFlag2SqlDSN(gameDB))
	sqlDSNs = append(sqlDSNs, mysql.AnalysisFlag2SqlDSN(memberDB))
	args.Sql = &frame.Sql{sqlDSNs}
	//框架启动完毕后执行的方法
	args.Loaded = start
	//通过参数启动框架
	frame.Load(args)
}

func start() {
	engine := eng.NewEngine(*addr)
	//将引擎注入到各模块
	inject(engine)
	//裁判端定制
	judgment()
	//将服务绑定到redis
	bindRedis()
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

//将服务绑定到redis
func bindRedis() {
	rds := frame.GetRedis("game")
	defer rds.Close()
	servers, err := redis_.Strings(rds.Do("lrange", "servers", 0, -1))
	logger.CheckError(err)
	//已绑定
	bound := false
	for _, server := range servers {
		if server == *addr {
			bound = true
		}
	}
	//未绑定
	if !bound {
		//初始化服务器的信息
		rds.Do("rpush", "servers", *addr)
		rds.Do("hmset", *addr, "pcount", 0, "matchpcount", "0,0,0,0,0", "roompcount", "0|0|0|0|0")
	}
}

//将服务从redis解除绑定
func unBindRedis() {
	rds := frame.GetRedis("game")
	defer rds.Close()
	servers, err := redis_.Strings(rds.Do("lrange", "servers", 0, -1))
	logger.CheckError(err)
	for _, server := range servers {
		if server == *addr {
			rds.Do("lrem", "servers", 0, *addr)
		}
	}
	rds.Do("expire", *addr, 0)
}

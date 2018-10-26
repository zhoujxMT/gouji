package main

import (
	"flag"

	"kelei.com/invest/cmds"
	. "kelei.com/invest/modes"
	"kelei.com/utils/frame"
	"kelei.com/utils/mysql"
	"kelei.com/utils/redis"
)

var (
	addr = flag.String("addr", "127.0.0.1:13000", "server address")
	//game,root,root2810,172.31.184.148:1052,investment
	db = flag.String("db", "game,root,111111,127.0.0.1:3306,investment", "")
)

func main() {
	//解析参数
	flag.Parse()
	//启动服务
	args := frame.Args{}
	args.ServerName = "invest"
	args.Commands = cmds.GetCmds()
	args.HttpGin = &frame.HttpGin{*addr, "", loadModes()}

	redisDSNs := []*redis.RedisDSN{}
	redisDSNs = append(redisDSNs, &redis.RedisDSN{"RDSName", "127.0.0.1:10100", ""})
	args.Redis = &frame.Redis{redisDSNs}

	sqlDSNs := []*mysql.SqlDSN{}
	sqlDSNs = append(sqlDSNs, mysql.AnalysisFlag2SqlDSN(db))
	args.Sql = &frame.Sql{sqlDSNs}

	args.Loaded = start
	frame.Load(args)
}

func loadModes() frame.Modes {
	modes := frame.Modes{}
	modes["Enterprise"] = &Enterprise{}
	modes["EnterpriseList"] = &EnterpriseList{}
	modes["Profitability"] = &Profitability{}
	modes["SharePrice"] = &SharePrice{}
	return modes
}

func start() {
}

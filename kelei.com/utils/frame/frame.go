/*
功能描述:服务框架
创建日期:8/3/18
创建人:hou
*/

package frame

import (
	"fmt"
)

const (
	Discovery_Peer2Peer = iota
	Discovery_Etcd
)

const (
	UNIDIRECTIONAL = iota //单向
	BIDIRECTIONAL         //双向
)

var (
	args Args
)

type Args struct {
	ServerName string                  //服务名称
	Commands   map[string]func(string) //命令块
	RpcxServer *RpcxServer             //rpcx服务器
	RpcxClient *RpcxClient             //rpcx客户端
	HttpGin    *HttpGin                //web http服务
	WebSocket  *WebSocket              //web socket服务
	Redis      *Redis                  //redis
	Sql        *Sql                    //sql
	Loaded     func()                  //加载框架完成
}

/*
1. p2p 只用第一个参数
2. etcd 客户端不用第一个参数
*/
type Rpcx struct {
	Addr        *string
	ForeignAddr *string
	EtcdAddr    *string
	BasePath    *string
}

//载入
func Load(args_ Args) {
	//启动服务
	fmt.Printf("[启动服务 %s]\n", args_.ServerName)
	//初始化参数列表
	initArgs(args_)
	//日志模块
	loadLog()
	//命令模块
	loadCommand()
	//rpcx模块
	loadRpcx()
	//http-gin模块
	loadHttpGin()
	//websocket模块
	loadWebSocket()
	//sql模块
	loadSql()
	//redis模块
	loadRedis()
	//框架加载完成
	loaded()
	//监听输入
	listenStdin()
}

//rpcx模块
func loadRpcx() {
	if args.RpcxServer != nil {
		loadRpcxServer()
	}
}

//框架加载完成,调用具体服务的内容
func loaded() {
	args.Loaded()
}

//设置args
func SetArgs(v Args) {
	args = v
}

//获取args
func GetArgs() Args {
	return args
}

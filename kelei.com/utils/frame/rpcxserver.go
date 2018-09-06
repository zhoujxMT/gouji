package frame

import (
	"time"

	metrics "github.com/rcrowley/go-metrics"
	"github.com/smallnest/rpcx/server"
	"github.com/smallnest/rpcx/serverplugin"

	"kelei.com/utils/logger"
)

var (
	rpcxServer *server.Server
)

type RpcxServer struct {
	DiscoveryType int
	Rpcx
	Modules []interface{}
}

func loadRpcxServer() {
	rpcx := args.RpcxServer
	rpcxServer = server.NewServer()
	if rpcx.DiscoveryType == Discovery_Etcd {
		addRegistryPlugin()
	}
	for _, module := range rpcx.Modules {
		rpcxServer.Register(module, "")
	}
	go rpcxServer.Serve("tcp", *rpcx.Addr)
}

//etcd专用
func addRegistryPlugin() {
	rpcx := args.RpcxServer
	r := &serverplugin.EtcdRegisterPlugin{
		ServiceAddress: "tcp@" + *rpcx.ForeignAddr,
		EtcdServers:    []string{*rpcx.EtcdAddr},
		BasePath:       *rpcx.BasePath,
		Metrics:        metrics.NewRegistry(),
		UpdateInterval: time.Minute,
	}
	err := r.Start()
	if err != nil {
		logger.Errorf(err.Error())
		return
	}
	rpcxServer.Plugins.Add(r)
}

func GetRpcxServer() *server.Server {
	return rpcxServer
}

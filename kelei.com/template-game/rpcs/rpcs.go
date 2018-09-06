package rpcs

import (
	"context"
	"net"
	"reflect"

	"github.com/smallnest/rpcx/server"

	eng "kelei.com/richman/engine"

	"kelei.com/utils/logger"
	"kelei.com/utils/rpcs"
)

var (
	engine *eng.Engine
)

func Inject(engine_ *eng.Engine) {
	engine = engine_
}

type RichManS struct {
}

func (this *RichManS) Call(ctx context.Context, args *rpcs.Args, reply *rpcs.Reply) error {
	clientConn := ctx.Value(server.RemoteConnContextKey).(net.Conn)
	args.V["clientConn"] = clientConn
	v := reflect.ValueOf(engine)
	funcName := args.V["funcname"].(string)
	logger.Debugf("收到请求:%s%v", funcName, args.V["args"])
	mv := v.MethodByName(funcName)
	res := mv.Call([]reflect.Value{reflect.ValueOf(args)})
	reply.V = res[0].Interface().(*string)
	logger.Debugf("回发结果:%s", *reply.V)
	return nil
}

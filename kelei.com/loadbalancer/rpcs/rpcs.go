package rpcs

import (
	"context"

	"kelei.com/utils/common"
	"kelei.com/utils/frame"
	"kelei.com/utils/logger"
	. "kelei.com/utils/rpcs"
)

type LoadbalancerS struct {
}

func (this *LoadbalancerS) GetGateAddr(ctx context.Context, args *Args, reply *Reply) error {
	reply.V = this.getGateAddr()
	return nil
}

func (this *LoadbalancerS) getGateAddr() *string {
	xclient := frame.NewRpcxClient(frame.GetArgs().RpcxClient)
	defer xclient.Close()
	args := &Args{}
	args.V = make(map[string]interface{})
	reply := &Reply{}
	err := xclient.Call(context.Background(), "GetGateAddr", args, reply)
	if err != nil {
		logger.Errorf("failed to call: %s", err.Error())
		return &common.STATUS_CODE_UNKNOWN
	}
	return reply.V
}

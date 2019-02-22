package rpcs

import (
	"context"
	"fmt"

	"kelei.com/utils/common"
	"kelei.com/utils/frame"
	. "kelei.com/utils/rpcs"
)

type GateS struct {
}

func (this *GateS) GetGateAddr(ctx context.Context, args *Args, reply *Reply) error {
	reply.RS = &frame.GetArgs().WebSocket.ForeignAddr
	fmt.Println("xxxxx:", frame.GetArgs().WebSocket.ForeignAddr)
	reply.SC = common.SC_OK
	return nil
}

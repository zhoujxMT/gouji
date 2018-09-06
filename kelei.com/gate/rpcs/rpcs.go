package rpcs

import (
	"context"

	"kelei.com/utils/frame"
	. "kelei.com/utils/rpcs"
)

type GateS struct {
}

func (this *GateS) GetGateAddr(ctx context.Context, args *Args, reply *Reply) error {
	reply.V = &frame.GetArgs().WebSocket.ForeignAddr
	return nil
}

package modes

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"kelei.com/utils/common"
	"kelei.com/utils/frame"
	"kelei.com/utils/rpcs"
)

type Login struct{}

type Result_Login struct {
	UserID   string `json:"userid"`
	GateAddr string `json:"gateaddr"`
}

func (this *Login) Login(c *gin.Context, args []string) {
	uid := args[0]
	userid := uid
	gateAddr := this.getGateAddr()
	res := Result_Login{userid, *gateAddr}
	c.JSON(http.StatusOK, res)
}

func (this *Login) getGateAddr() *string {
	xclient := frame.NewRpcxClient(frame.GetArgs().RpcxClient)
	defer xclient.Close()
	args := &rpcs.Args{}
	args.V = make(map[string]interface{})
	reply := &rpcs.Reply{}
	err := xclient.Call(context.Background(), "GetGateAddr", args, reply)
	if err != nil {
		fmt.Println("failed to call: ", err)
		return &common.STATUS_CODE_UNKNOWN
	}
	return reply.V
}

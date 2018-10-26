package modes

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"kelei.com/utils/common"
	"kelei.com/utils/frame"
	"kelei.com/utils/logger"
	"kelei.com/utils/rpcs"
)

const (
	CATEGORY_WEB = iota //网页
	CATEGORY_WX         //微信
)

type Login struct{}

type Result_Login struct {
	UID        string `json:"uid"`
	GateAddr   string `json:"gateaddr"`
	SessionKey string `json:"sessionkey"`
}

/*
	登录
	in:uid
	out:已连接服务器!
*/
func (this *Login) Login(c *gin.Context, args []string) {
	uid, sessionKey := "", ""
	category := common.ParseInt(args[1])
	if category == CATEGORY_WX {
		jscode := args[0]
		var err error
		uid, sessionKey, err = this.getOpenID(jscode)
		if err != nil {
			res := Result_Login{}
			c.JSON(http.StatusOK, res)
			return
		}
	} else if category == CATEGORY_WEB {
		uid = args[0]
	}
	this.register(uid)
	gateAddr := this.getGateAddr()
	res := Result_Login{uid, *gateAddr, sessionKey}
	c.JSON(http.StatusOK, res)
}

func (this *Login) register(uid string) {
	db := frame.GetDB("member")
	dbuid := ""
	rows, err := db.Query("select uid from memberinfo where uid=?", uid)
	logger.CheckFatal(err, "register")
	defer rows.Close()
	for rows.Next() {
		rows.Scan(&dbuid)
	}
	//没有此玩家,注册
	if dbuid == "" {
		rows, err := db.Query("insert into memberinfo(uid,password) values(?,?)", uid, 123)
		logger.CheckFatal(err, "register2")
		defer rows.Close()
	}
}

func (this *Login) getGateAddr() *string {
	xclient := frame.NewRpcxClient(frame.GetArgs().RpcxClient)
	defer xclient.Close()
	args := &rpcs.Args{}
	reply := &rpcs.Reply{}
	err := xclient.Call(context.Background(), "GetGateAddr", args, reply)
	if err != nil {
		fmt.Println("failed to call: ", err)
		return &common.SC_LBERR
	}
	return reply.RS
}

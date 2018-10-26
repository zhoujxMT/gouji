package frame

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"

	. "kelei.com/utils/common"
	"kelei.com/utils/logger"
)

var (
	modes = make(Modes)
)

type HttpGin struct {
	Addr        string
	Certificate string
	Modes       Modes
}

type Modes map[string]interface{}

//gin模块
func loadHttpGin() {
	if args.HttpGin == nil {
		return
	}
	addr := args.HttpGin.Addr
	logger.Infof("[启动http %s]", addr)
	modes = args.HttpGin.Modes
	gin.SetMode(gin.ReleaseMode)
	//开启路由器
	router := gin.Default()
	//注册接口
	router.POST("/"+args.ServerName, posting)
	//监听端口
	address := NewAddress(addr)
	//有证书,就启用https
	if args.HttpGin.Certificate != "" {
		go func() {
			err := router.RunTLS(":"+address.Port, args.HttpGin.Certificate+".pem", args.HttpGin.Certificate+".key")
			logger.CheckFatal(err)
		}()
	}
	go http.ListenAndServe(fmt.Sprintf(":%s", address.Port), router)
}

//获取类模型
func getMode(modeName string) (reflect.Value, error) {
	var v reflect.Value
	var err error
	mode := modes[modeName]
	if mode == nil {
		err = errors.New("类不存在")
		return v, err
	}
	t := reflect.ValueOf(mode).Type()
	v = reflect.New(t).Elem()
	return v, nil
}

//post请求
func posting(c *gin.Context) {
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*") //允许访问所有域
	modeName := c.Request.FormValue("mode")
	funcName := c.Request.FormValue("func")
	args_s := c.Request.FormValue("args")
	args := []string{}
	if args_s != "" {
		args = strings.Split(args_s, ",")
	}
	mode, err := getMode(modeName)
	if err == nil {
		method := mode.MethodByName(funcName)
		method.Call([]reflect.Value{reflect.ValueOf(c), reflect.ValueOf(args)})
	}
}

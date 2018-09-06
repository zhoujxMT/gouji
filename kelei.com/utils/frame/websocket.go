package frame

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/smallnest/rpcx/client"

	"kelei.com/utils/common"
	"kelei.com/utils/logger"
	"kelei.com/utils/rpcs"
	websocket_ "kelei.com/utils/websocket"
)

const (
	WSType_wss = iota
	WSType_ws
)

var (
	clientManager = &ClientManager{make(map[string]*Client)}
)

func GetClientManager() *ClientManager {
	return clientManager
}

type WebSocket struct {
	Addr             string
	ForeignAddr      string
	WSType           int
	Init             func(uid string) (string, error)
	HandleClientData func(*rpcs.Args) (string, bool)
}

func loadWebSocket() {
	if args.WebSocket == nil {
		return
	}
	ws := websocket_.New(args.WebSocket.Addr, args.WebSocket.WSType)
	ws.Start(handleConn)
}

func handleConn(conn *websocket.Conn) {
	clientManager.AddClient(conn)
}

/*
================================================================================
================================================================================
================================================================================
*/

type ClientManager struct {
	clients map[string]*Client
}

//创建客户端
func (c *ClientManager) createClient(conn *websocket.Conn) *Client {
	client := &Client{}
	client.conn = conn
	client.sendMsg = make(chan []byte)
	client.lastActionTime = time.Now()
	go client.read()
	go client.write()
	return client
}

//添加客户端 websocket
func (c *ClientManager) AddClient(conn *websocket.Conn) {
	addr := conn.RemoteAddr().String()
	logger.Debugf("客户端<%s>连接成功", addr)
	//用户连接成功
	client := c.createClient(conn)
	//关闭连接机制
	client.closeConnMechanism()
	//向客户端发送连接成功
	client.send([]byte("0ConnectSucess,已连接服务器!"))
}

//保存客户端
func (c *ClientManager) saveClient(client *Client) {
	c.clients[client.getUserID()] = client
}

//删除客户端
func (c *ClientManager) removeClient(userid string) {
	delete(c.clients, userid)
}

//获取客户端
func (c *ClientManager) GetClient(userid string) *Client {
	return c.clients[userid]
}

/*
================================================================================
================================================================================
================================================================================
*/

type Client struct {
	uid            string          //平台userid
	userid         string          //游戏userid
	headUrl        string          //玩家头像
	userName       string          //玩家名字
	conn           *websocket.Conn //连接
	rpcClient      client.XClient  //
	sendMsg        chan []byte     //发送给客户端的信息
	lastActionTime time.Time       //最后活动的时间
}

func (c *Client) getUID() string {
	return c.uid
}

func (c *Client) setUID(uid string) {
	c.uid = uid
}

func (c *Client) getUserID() string {
	return c.userid
}

func (c *Client) setUserID(userid string) {
	c.userid = userid
}

func (c *Client) GetRpcxClient() client.XClient {
	return c.rpcClient
}

func (c *Client) closeRpcxClient() {
	c.rpcClient.Close()
}

//设置最后活动时间
func (c *Client) setLastActionTime(t time.Time) {
	c.lastActionTime = t
}

//向客户端发送数据带打印
func (c *Client) Send_log(msg []byte) {
	c.send_log(msg)
}

//向客户端发送数据带打印
func (c *Client) send_log(msg []byte) {
	logger.Debugf("客户端<%s>返回结果:%s", c.conn.RemoteAddr(), string(msg))
	c.send(msg)
}

//向客户端发送数据
func (c *Client) send(msg []byte) {
	c.sendMsg <- msg
}

//关闭客户端
func (c *Client) close(err error) {
	//从客户端列表中删除
	clientManager.removeClient(c.getUserID())
	//向游戏服务器发送连接关闭
	c.handleClientData([]byte("0ClientConnClose"))
	//关闭rpcx连接
	c.closeRpcxClient()
	logger.Debugf("客户端 断开 : %s", err)
}

//读取连接
func (c *Client) read() {
	defer func() {
		c.conn.Close()
	}()

	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			c.close(err)
			break
		}
		c.setLastActionTime(time.Now())
		c.handleClientData(msg)
	}
}

//写入连接
func (c *Client) write() {
	defer func() {
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.sendMsg:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.conn.WriteMessage(websocket.TextMessage, message)
		}
	}
}

/*
关闭连接机制
连接成功后1秒钟没进行初始化,就当作机器人,进行关闭
*/
func (c *Client) closeConnMechanism() {
	go func() {
		time.Sleep(time.Second * 1)
		if c.getUID() == "" {
			c.closeConn()
		}
	}()
}

//关闭客户端
func (c *Client) closeConn() {
	c.conn.Close()
}

//连接游戏服
func (c *Client) connect() {
	//连接游戏服rpcserver
	c.rpcClient = NewRpcxClient(args.RpcxClient)
	//连接
	content := "0Connect"
	c.handleClientData([]byte(content))
}

//处理客户端数据
func (c *Client) handleClientData(msg []byte) string {
	//	defer func() {
	//		if p := recover(); p != nil {
	//			logger.Errorf("[recovery] handleClientData : %v", p)
	//		}
	//	}()
	frameArgs := args
	rpcArgs, funcIndex := c.ParseData(msg)
	funcName := rpcArgs.V["funcname"].(string)
	push := func(message string) {
		c.send([]byte(fmt.Sprintf("%d%s,%s", funcIndex, funcName, message)))
	}
	push = push
	log := func(message string) {
		logger.Debugf("客户端<%s>返回结果:%d%s,%s", c.conn.RemoteAddr(), 0, funcName, string(message))
	}
	log = log
	push_log := func(message string) {
		c.send_log([]byte(fmt.Sprintf("%d%s,%s", funcIndex, funcName, message)))
	}
	if funcName != "PING" {
		logger.Debugf("客户端<%s>命令调用:%s", c.conn.RemoteAddr(), string(msg))
	}
	//客户端初始化,连接“游戏服务器”
	if funcName == "Init" {
		p := rpcArgs.V["args"].([]string)
		uid := p[1]
		userid, err := frameArgs.WebSocket.Init(uid)
		if err == nil {
			c.setUID(uid)
			c.setUserID(userid)
			clientManager.saveClient(c)
			c.connect()
			push_log(userid)
		} else {
			c.closeConn()
		}
		return ""
	}
	//初始化没有成功,关闭连接
	if c.getUID() == "" {
		c.closeConn()
		return ""
	}
	if funcName == "PING" { //心跳包,用来断线重连
		push("PONG")
		return ""
	}
	//处理用户自定义方法
	if frameArgs.WebSocket.HandleClientData != nil {
		res, termination := frameArgs.WebSocket.HandleClientData(rpcArgs)
		if termination {
			push_log(res)
			return ""
		}
	}
	//调用游戏服
	res := *c.Call(rpcArgs)
	//回发到客户端
	if res != common.STATUS_CODE_NOBACK {
		ispush := c.rpcCallBack(rpcArgs)
		if ispush {
			push_log(res)
		} else {
			log(res)
		}
	}
	return ""
}

//调用游戏服
func (c *Client) Call(args *rpcs.Args) *string {
	xclient := c.GetRpcxClient()
	reply := &rpcs.Reply{}
	err := xclient.Call(context.Background(), "Call", args, reply)
	if err != nil {
		logger.Errorf("failed to call: %s", err.Error())
		return &common.STATUS_CODE_UNKNOWN
	}
	return reply.V
}

//调用游戏服回调
func (c *Client) rpcCallBack(rpcArgs *rpcs.Args) bool {
	funcName := rpcArgs.V["funcname"]
	ispush := true
	//退出比赛,切换到“游戏服务器”
	if funcName == "Connect" {
		ispush = false
	}
	return ispush
}

//解析数据
func (c *Client) ParseData(msg []byte) (*rpcs.Args, int) {
	data := strings.Split(string(msg), ",")
	funcIndex, err := strconv.Atoi(data[0][:1])
	logger.CheckFatal(err)
	funcIndex = common.Clampf(funcIndex, 1, 9)
	funcName := data[0][1:]
	v := make(map[string]interface{})
	v["funcname"] = funcName
	p := data[1:]
	v["args"] = common.InsertStringSlice(p, []string{c.getUserID()}, 0)
	args := rpcs.Args{v}
	return &args, funcIndex
}

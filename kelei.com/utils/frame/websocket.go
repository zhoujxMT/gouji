package frame

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/garyburd/redigo/redis"
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
	clientManager = &ClientManager{clients: make(map[string]*Client), clientsName: make(map[string]*string)}
)

func GetClientManager() *ClientManager {
	return clientManager
}

type WebSocket struct {
	Addr             string
	ForeignAddr      string
	Certificate      string
	HandleClientData func(string, []string) (string, bool)
}

func loadWebSocket() {
	if args.WebSocket == nil {
		return
	}
	ws := websocket_.New(args.WebSocket.Addr, args.WebSocket.Certificate)
	ws.Start(handleConn)
	initSDK()
}

func handleConn(conn *websocket.Conn) {
	if clientManager.serverStarted() {
		clientManager.AddClient(conn)
	} else {
		time.Sleep(time.Second * 5)
		conn.Close()
	}
}

/*
================================================================================
================================================================================
================================================================================
*/

type ClientManager struct {
	clients     map[string]*Client
	clientsName map[string]*string
	lock        sync.Mutex
}

//服务开启完毕
func (c *ClientManager) serverStarted() bool {
	return GetDB() != nil
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
	/*
		向客户端发送连接成功
		push:已连接服务器!
	*/
	message := client.Format("ConnectSucess", "已连接服务器!")
	client.send([]byte(message))
}

//保存客户端
func (c *ClientManager) saveClient(client *Client) {
	c.lock.Lock()
	cl := c.clients[client.getUserID()]
	c.lock.Unlock()
	if cl != nil {
		err := errors.New("建立新连接,关闭旧连接")
		cl.close(err)
		//只有裁判端才有用
		client.Log_Push(client.Format("Pause_Push", "1"))
	}
	c.lock.Lock()
	c.clients[client.getUserID()] = client
	c.lock.Unlock()
}

//删除客户端
func (c *ClientManager) removeClient(userid string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	delete(c.clients, userid)
}

//获取客户端
func (c *ClientManager) GetClient(userid string) *Client {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.clients[userid]
}

//获取客户端根据uid
func (c *ClientManager) GetClientByUID(uid string) *Client {
	c.lock.Lock()
	defer c.lock.Unlock()
	for _, cl := range c.clients {
		if cl.getUID() == uid {
			return cl
		}
	}
	return nil
}

//获取玩家名字
func (c *ClientManager) getUserName(uid string) *string {
	return c.clientsName[uid]
}

//预保存玩家名字
func (c *ClientManager) saveUserName(uid string, userName *string) {
	c.clientsName[uid] = userName
}

/*
================================================================================
================================================================================
================================================================================
*/

type Client struct {
	uid             string          //平台userid
	userid          string          //游戏userid
	headUrl         string          //玩家头像
	userName        string          //玩家名字
	sessionKey      string          //
	conn            *websocket.Conn //连接
	rpcClient       client.XClient  //
	sendMsg         chan []byte     //发送给客户端的信息
	lastActionTime  time.Time       //最后活动的时间
	currentRoomID   string          //当前所在房间id
	currentRoomInfo string          //当前房间的信息
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

func (c *Client) getUserName() string {
	return c.userName
}

func (c *Client) setUserName(userName string) {
	c.userName = userName
}

func (c *Client) getCurrentRoomID() string {
	return c.currentRoomID
}

func (c *Client) setCurrentRoomID(currentRoomID string) {
	c.currentRoomID = currentRoomID
}

func (c *Client) getCurrentRoomInfo() string {
	return c.currentRoomInfo
}

func (c *Client) setCurrentRoomInfo(currentRoomInfo string) {
	c.currentRoomInfo = currentRoomInfo
}

func (c *Client) getSessionKey() string {
	return c.sessionKey
}

func (c *Client) setSessionKey(sessionKey string) {
	c.sessionKey = sessionKey
}

func (c *Client) getHeadUrl() string {
	return c.headUrl
}

func (c *Client) setHeadUrl(headUrl string) {
	c.headUrl = headUrl
}

func (c *Client) GetRpcxClient() client.XClient {
	return c.rpcClient
}

func (c *Client) closeRpcxClient() {
	if c.rpcClient != nil {
		c.rpcClient.Close()
	}
}

//设置最后活动时间
func (c *Client) setLastActionTime(t time.Time) {
	c.lastActionTime = t
}

//关闭客户端
func (c *Client) close(err error) {
	c.conn.Close()
	//从客户端列表中删除
	clientManager.removeClient(c.getUserID())
	//向游戏服务器发送连接关闭
	c.handleClientData([]byte("ClientConnClose"))
	//关闭rpcx连接
	c.closeRpcxClient()
	logger.Debugf("客户端<%s>断开 : %s", c.conn.RemoteAddr(), err)
}

//读取连接
func (c *Client) read() {
	defer func() {
		c.conn.Close()
		if p := recover(); p != nil {
			logger.Errorf("[recovery] read : %v", p)
		}
	}()
	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			errStr := err.Error()
			logger.Errorf("websocket err:%s", errStr)
			if strings.Contains(errStr, "1000") || strings.Contains(errStr, "1001") || strings.Contains(errStr, "connection reset by peer") || strings.Contains(errStr, "repeated read on failed websocket connection") || strings.Contains(errStr, "use of closed network connection") {
				c.close(err)
				break
			}
		} else {
			c.setLastActionTime(time.Now())
			c.handleClientData(msg)
		}
	}
}

//写入连接
func (c *Client) write() {
	defer func() {
		c.conn.Close()
		if p := recover(); p != nil {
			logger.Errorf("[recovery] write : %v", p)
		}
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
		defer func() {
			if p := recover(); p != nil {
				logger.Errorf("[recovery] closeConnMechanism : %v", p)
			}
		}()
		//服务器已启动
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

//连接游戏服(etcd)
func (c *Client) connect() {
	//连接游戏服rpcserver
	c.rpcClient = NewRpcxClient(args.RpcxClient)
	//连接
	content := "Connect"
	c.handleClientData([]byte(content))
}

//连接游戏服(Peer2Peer)
func (c *Client) connectP2P(rpcxClient *RpcxClient) {
	//连接游戏服rpcserver
	c.rpcClient = NewRpcxClient(rpcxClient)
	//连接
	content := "Connect"
	c.handleClientData([]byte(content))
}

//向客户端发送数据
func (c *Client) send(msg []byte) {
	c.sendMsg <- msg
}

type Result struct {
	FN string //方法名
	RS string //返回值
	SC int    //状态码
}

//格式化
func (c *Client) Format(funcName string, message string, args ...string) string {
	statusCode := common.SC_OK
	if len(args) >= 1 {
		statusCode = args[0]
	}
	result := Result{funcName, message, common.ParseInt(statusCode)}
	b, err := json.Marshal(result)
	logger.CheckError(err)
	msg := string(b)
	//+++
	msg = fmt.Sprintf("%s,%s", funcName, message)
	return msg
}

//压缩json
func (c *Client) compress(str *string) *string {
	newStr := *str
	return &newStr
}

//压缩json
//func (c *Client) compress(json *string) *string {
//	pat := "\"[a-zA-Z]*?\":"
//	if ok, _ := regexp.Match(pat, []byte(*json)); ok {
//		fmt.Println("match found")
//	}
//	f := func(s string) string {
//		fmt.Println("aaa:", s)
//		s = strings.Replace(s, "\\\"", "'", 2)
//		return s
//	}
//	re, _ := regexp.Compile(pat)
//	newStr := re.ReplaceAllStringFunc(*json, f)
//	fmt.Println(newStr)
//	return &newStr
//}

func (c *Client) log(message string) {
	logger.Debugf("客户端<%s>返回结果<%d>:%s", c.conn.RemoteAddr(), len(message), message)
}

func (c *Client) push(message string) {
	c.send([]byte(message))
}

func (c *Client) Log_Push(message string) {
	c.log(message)
	c.push(message)
}

//初始化,获取userid
func (c *Client) Init(uid, username string) (userid string, err error) {
	if c.checkUID(uid) {
		userid = c.createUser(uid, username)
	} else {
		err = errors.New("无效的uid")
	}
	return userid, err
}

//检查userid是否存在
func (c *Client) checkUID(uid string) bool {
	db := GetDB("member")
	count := 0
	err := db.QueryRow("select count(*) from memberinfo where uid=?", uid).Scan(&count)
	logger.CheckError(err, "checkUID")
	if count == 0 {
		logger.Warnf("无效的uid")
		return false
	}
	return true
}

/*
创建玩家,如果玩家不存在就创建
in:
out:userid
*/
func (c *Client) createUser(uid, username string) string {
	res := ""
	db := GetDB("game")
	userid := -1
	db.QueryRow("select userinfo.userid from userinfo where userinfo.uid=?", uid).Scan(&userid)
	if userid == -1 {
		_, err := db.Exec("insert into userinfo(uid,username) values(?,?)", uid, username)
		logger.CheckFatal(err, "CreateUser")
		err = db.QueryRow("select userid from userinfo where uid=?", uid).Scan(&userid)
		logger.CheckFatal(err, "CreateUser2")
		_, err = db.Exec("call UserInit(?)", userid)
		logger.CheckFatal(err, "CreateUser3")
	} else {
		db.Exec("update userinfo set userinfo.username=? where uid=?", username, uid)
	}
	res = strconv.Itoa(userid)
	return res
}

//获取玩家基本信息(玩家必须在线)
func (c *Client) GetBaseInfo(userid string) string {
	cl := clientManager.GetClient(userid)
	if cl == nil {
		return ""
	}
	return userid + "," + cl.getHeadUrl() + "," + cl.getUserName()
}

//获取玩家基本信息(玩家不在线也可以)
func (c *Client) GetBaseInfoByUID(uid string) string {
	p := clientManager.getUserName(uid)
	userName := ""
	if p != nil {
		userName = *p
	} else {
		userName = c.getUserNameFromDB(uid)
	}
	return uid + ",," + userName
}

/*
是否有正在进行的比赛
*/
func (c *Client) ExistingMatch() {
	gameaddr, roomid, matchid, roomtype := "", "-1", "", ""
	key := fmt.Sprintf("usermatch:%s", c.getUserID())
	rds := GetRedis("game")
	defer rds.Close()
	data, err := redis.Strings(rds.Do("hmget", key, "gameaddr", "roomid", "matchid", "roomtype"))
	logger.CheckError(err)
	if data[0] != "" {
		gameaddr, roomid, matchid, roomtype = data[0], data[1], data[2], data[3]
	}
	res := "0"
	//有正在进行的比赛
	if roomid != "-1" {
		fmt.Println("有正在进行的比赛")
		res = fmt.Sprintf("%s|%s", matchid, roomtype)
		module := args.RpcxClient.Module
		c.connectP2P(&RpcxClient{Discovery_Peer2Peer, Rpcx{&gameaddr, nil, nil, nil}, module, BIDIRECTIONAL})
	} else {
		fmt.Println("没有正在进行的比赛")
		c.connect()
	}
	c.setCurrentRoomID(roomid)
	c.setCurrentRoomInfo(res)
}

//获取username
func (c *Client) getUserNameFromDB(uid string) string {
	db := GetDB("game")
	userName := ""
	err := db.QueryRow("select videousername from userinfo where uid=?", uid).Scan(&userName)
	logger.CheckError(err)
	if userName != "" {
		clientManager.saveUserName(uid, &userName)
	}
	return userName
}

//处理客户端数据
func (c *Client) handleClientData(msg []byte) string {
	if GetMode() == MODE_RELEASE {
		defer func() {
			if p := recover(); p != nil {
				logger.Errorf("[recovery] handleClientData : %v", p)
			}
		}()
	}
	frameArgs := args
	rpcArgs := c.ParseData(msg)
	funcName := rpcArgs.V["funcname"].(string)
	p := rpcArgs.V["args"].([]string)
	if funcName != "PING" {
		logger.Debugf("客户端<%s>命令调用:%s", c.conn.RemoteAddr(), string(msg))
	}
	/*
		客户端初始化,连接“游戏服务器”
		in:uid,headUrl,userName
		out:userid
	*/
	if funcName == "Init" {
		uid := p[1]
		headUrl := p[2]
		userName := p[3]
		sessionKey := p[4]
		userid, err := c.Init(uid, userName)
		if err == nil {
			c.setUserID(userid)
			c.setHeadUrl(headUrl)
			c.setUserName(userName)
			c.setSessionKey(sessionKey)
			c.setUID(uid)
			c.setUserID(userid)
			clientManager.saveClient(c)
			c.ExistingMatch()
			name := clientManager.getUserName(uid)
			if name != nil {
				c.setUserName(*name)
			}
			c.Log_Push(c.Format(funcName, userid))
		} else {
			c.closeConn()
		}
		return ""
	}
	//初始化没有成功,关闭连接
	if c.getUID() == "" {
		return ""
	}
	if funcName == "PING" { //心跳包,用来断线重连
		c.push(c.Format(funcName, "PONG"))
		return ""
	}
	if funcName == "GetBaseInfo" { //
		userid := p[1]
		c.Log_Push(c.Format(funcName, c.GetBaseInfo(userid)))
		return ""
	}
	if funcName == "GetBaseInfoByUID" { //
		uid := p[1]
		c.Log_Push(c.Format(funcName, c.GetBaseInfoByUID(uid)))
		return ""
	}
	if funcName == "GetToken" { //心跳包,用来断线重连
		c.push(c.Format(funcName, token))
		return ""
	}
	if funcName == "GetRoomQRCode" { //获取房间的二维码 out:-101不在比赛中 其它图片的url
		c.Log_Push(c.Format(funcName, c.getRoomQRCode()))
		return ""
	}
	if funcName == "UpdateUserName" { //修改玩家名
		uid := p[1]
		newUserName := p[2]
		clientManager.saveUserName(uid, &newUserName)
		db := GetDB("game")
		_, err := db.Exec("update userinfo set videousername=? where uid=?", newUserName, uid)
		logger.CheckError(err)
		cl := clientManager.GetClientByUID(uid)
		if cl != nil {
			cl.setUserName(newUserName)
		}
		c.push(c.Format(funcName, common.Res_Succeed))
		clientManager.lock.Lock()
		for _, cl := range clientManager.clients {
			cl.push(c.Format(funcName, fmt.Sprintf("%s,%s", uid, newUserName)))
		}
		clientManager.lock.Unlock()
		return ""
	}
	if funcName == "ExistingMatch" { //是否有正在进行的比赛
		res := c.getCurrentRoomInfo()
		c.Log_Push(c.Format(funcName, res))
		return ""
	}
	if funcName == "Connect" { //连接游戏服务器
		additional := []string{c.getUID(), token, c.getSessionKey(), Secret, c.getHeadUrl()}
		p = append(p, additional...)
		rpcArgs.V["args"] = p
	}
	if funcName == "Matching" { //连接游戏服务器
		additional := []string{c.getCurrentRoomID()}
		p = append(p, additional...)
		rpcArgs.V["args"] = p
	}
	if funcName == "EnterRoom" { //连接游戏服务器
		additional := []string{c.getCurrentRoomID()}
		p = append(p, additional...)
		rpcArgs.V["args"] = p
	}
	//处理用户自定义方法
	if frameArgs.WebSocket.HandleClientData != nil {
		res, termination := frameArgs.WebSocket.HandleClientData(funcName, p)
		if termination {
			c.Log_Push(c.Format(funcName, res))
			return ""
		} else {
			if res != "" {
				additional := strings.Split(res, ",")
				p = append(p, additional...)
			}
		}
	}
	//调用游戏服
	reply := *c.Call(rpcArgs)
	//回发到客户端
	if reply.SC != common.SC_NOBACK {
		rs := ""
		if reply.RS != nil {
			rs = *reply.RS
		}
		ispush := c.rpcCallBack(rpcArgs)
		if ispush {
			c.Log_Push(c.Format(funcName, rs, reply.SC))
		} else {
			c.log(c.Format(funcName, rs, reply.SC))
		}
	}
	return ""
}

//调用游戏服
func (c *Client) Call(args *rpcs.Args) *rpcs.Reply {
	xclient := c.GetRpcxClient()
	reply := &rpcs.Reply{}
	err := xclient.Call(context.Background(), "Call", args, reply)
	if err != nil {
		logger.Errorf("failed to call: %s", err.Error())
		reply.SC = common.SC_GAMEERR
		return reply
	}
	return reply
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
func (c *Client) ParseData(msg []byte) *rpcs.Args {
	data := strings.Split(string(msg), ",")
	funcName := data[0]
	v := make(map[string]interface{})
	v["funcname"] = funcName
	p := data[1:]
	v["args"] = common.InsertStringSlice(p, []string{c.getUserID()}, 0)
	args := rpcs.Args{v}
	return &args
}

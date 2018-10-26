package websocket

import (
	"net/http"

	"github.com/gorilla/websocket"

	"kelei.com/utils/common"
	"kelei.com/utils/logger"
)

const (
	WSType_wss = iota
	WSType_ws
)

type WSServer struct {
	addr        string
	certificate string
	upgrader    websocket.Upgrader
}

func New(addr string, certificate string) *WSServer {
	wsServer := WSServer{}
	wsServer.addr = addr
	wsServer.certificate = certificate
	wsServer.upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	return &wsServer
}

func (this *WSServer) Start(handleConn func(conn *websocket.Conn)) {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := this.upgrader.Upgrade(w, r, nil)
		if err != nil {
			logger.Errorf("客户端连接失败: %s", err)
			return
		}
		go handleConn(conn)
	})
	logger.Infof("server started on : %s", this.addr)
	go func() {
		var err error
		if this.certificate == "" { //ws
			err = http.ListenAndServe(this.addr, nil)
		} else { //wss
			address := common.NewAddress(this.addr)
			err = http.ListenAndServeTLS(":"+address.Port, this.certificate+".pem", this.certificate+".key", nil)
		}
		logger.CheckFatal(err, "wsserver start")
	}()
}

//处理新连接 socket
//func (this *WSServer) handleConn(conn *websocket.Conn) {
//	logger.Debugf("yes")
//	//	client.ClientManage.AddClient(conn)
//}

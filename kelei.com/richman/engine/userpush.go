package engine

import (
	"fmt"
	"time"

	"kelei.com/utils/frame"
	"kelei.com/utils/logger"
)

func (this *User) push(funcName string, content *string) {
	go func() {
		time.Sleep(time.Millisecond)
		conn := this.getConn()
		xServer := frame.GetRpcxServer()
		msg := fmt.Sprintf("%s&%s&%s", this.getUserID(), funcName, *content)
		err := xServer.SendMessage(conn, "service_path", "service_method", nil, []byte(msg))
		logger.CheckError(err, fmt.Sprintf("failed to send messsage to %s : ", conn.RemoteAddr().String()))
	}()
}

func (this *User) pushGrid(grids []*Grid) {
	this.push("GridImage", this.getRoom().GetGridSystem().getImageChecked(grids))
}

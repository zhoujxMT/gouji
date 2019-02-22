/*
推送
*/

package push

import (
	"bytes"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"kelei.com/server/rpcs"
	. "kelei.com/utils/common"
	"kelei.com/utils/frame"
	"kelei.com/utils/logger"
)

var (
	userids              *string
	messages             *string
	db                   *sql.DB
	buff_userids_local   = bytes.Buffer{}
	buff_messages_local  = bytes.Buffer{}
	buff_userids_global  = bytes.Buffer{}
	buff_messages_global = bytes.Buffer{}
	lock                 sync.Mutex
)

func Init() {
	db = frame.GetDB("game")
	logger.Infof("[启动推送服务]")
	go func() {
		for {
			handle()
			time.Sleep(time.Second * 1)
		}
	}()
	go func() {
		for {
			broadcast()
			time.Sleep(time.Second * 1)
		}
	}()
}

//重置局部字符串
func resetLocalBuff() {
	buff_userids_local.Reset()
	buff_messages_local.Reset()
}

//重置全局字符串
func resetGlobalBuff() {
	buff_messages_global.Reset()
}

//重置字符串
func resetBuff() {
	resetLocalBuff()
	resetGlobalBuff()
}

//从数据库获取推送消息
func handle() {
	if frame.GetMode() == frame.MODE_RELEASE {
		defer func() {
			if p := recover(); p != nil {
				logger.Errorf("[recovery] handle : %v", p)
			}
		}()
	}
	lock.Lock()
	defer lock.Unlock()
	type_, userid, content := "", "", ""
	rows, err := db.Query("call PushGet()")
	logger.CheckFatal(err)
	defer rows.Close()
	//	resetBuff()
	for rows.Next() {
		rows.Scan(&type_, &userid, &content)
		if userid == "0" {
			buff_messages_global.WriteString(fmt.Sprintf("%s|%s$", type_, content))
		} else {
			buff_userids_local.WriteString(fmt.Sprintf("%s,", userid))
			buff_messages_local.WriteString(fmt.Sprintf("%s|%s^", type_, content))
		}
	}
}

/*
Push(推送所有的信息)
push:消息类型|消息内容$
*/
func broadcast() {
	if frame.GetMode() == frame.MODE_RELEASE {
		defer func() {
			if p := recover(); p != nil {
				logger.Errorf("[recovery] broadcast : %v", p)
			}
		}()
	}
	lock.Lock()
	defer lock.Unlock()
	messages_global := *RemoveLastChar(buff_messages_global)
	if messages_global != "" {
		rpcs.PushMessage("*", "Push", messages_global)
		resetGlobalBuff()
	}
	messages_local := *RemoveLastChar(buff_messages_local)
	if messages_local != "" {
		userids_local := *RemoveLastChar(buff_userids_local)
		rpcs.PushMessage(userids_local, "Push", messages_local)
		resetLocalBuff()
	}
	resetBuff()
}

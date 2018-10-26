package update

import (
	"database/sql"
	"fmt"
	"time"

	"kelei.com/utils/frame"
	"kelei.com/utils/logger"
)

type Event struct {
	Name           string //事件名
	TriggerWeekDay string //触发星期几
	TriggerTime    string //触发时间
	TriggerFunc    func() //触发方法
	Des            string //描述
}

/*
	Monday
	Tuesday
	Wednesday
	Thursday
	Friday
	Saturday
	Sunday
*/

var (
	events = []Event{
		Event{"resetRanking", "Wednesday", "10:27:10", resetRanking, "重置排名{event resetRanking,Wednesday,11:01:00}"},
	}
)

var (
	db *sql.DB
)

func SetEvent(name, TriggerWeekDay, TriggerTime string) {
	for i, event := range events {
		if event.Name == name {
			event.TriggerWeekDay = TriggerWeekDay
			event.TriggerTime = TriggerTime
			events[i] = event
		}
	}
}

func Init() {
	db = frame.GetDB("game")
	logger.Infof("[启动更新服务]")
	go func() {
		for {
			handle()
			time.Sleep(time.Second * 1)
		}
	}()
}

func handle() {
	if frame.GetMode() == frame.MODE_RELEASE {
		defer func() {
			if p := recover(); p != nil {
				logger.Errorf("[recovery] Server_Init : %v", p)
			}
		}()
	}
	timeNow := time.Now()
	nowWeekday := timeNow.Weekday().String()
	nowTime := timeNow.Format("15:04:05")
	logger.Debugf("当前时间：%s,%s", nowWeekday, nowTime)
	for _, event := range events {
		if event.TriggerWeekDay == nowWeekday && event.TriggerTime == nowTime {
			event.TriggerFunc()
		}
	}
}

/*
重置排行
*/
func resetRanking() {
	fmt.Println("重置排行")
	db.Exec("update ranking set win=0,charm=0")
}

package common

import (
	"strconv"
	"time"
)

const (
	ExpTime_Second = 1
	ExpTime_Short  = 10
	ExpTime_Minute = 60
	ExpTime_Hour   = 60 * 60
)

func ParseTime(time_s string) (time.Time, error) {
	t, err := time.ParseInLocation("2006-01-02 15:04:05", time_s, time.Local)
	return t, err
}

//获取时间戳
func GetTimestamp() string {
	cur := time.Now()
	timestamp := cur.UnixNano() / 1000000 //毫秒ranking
	str := strconv.FormatInt(timestamp, 10)
	str = str[len(str)-6:]
	return str
}

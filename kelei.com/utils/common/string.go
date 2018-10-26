package common

import (
	"bytes"
	"strconv"
)

//数据库行转化成buff之后,将buff转化成字符串,并检测是否没有数据
func RowsBufferToString(buff bytes.Buffer) *string {
	str := buff.String()
	if str != "" {
		str = str[:len(str)-1]
	}
	if str == "" {
		str = Res_NoData
	}
	return &str
}

//去掉最后一个多余的字符
func RemoveLastChar(buff bytes.Buffer) *string {
	str := buff.String()
	if str != "" {
		str = str[:len(str)-1]
	}
	return &str
}

//字符串转化成数字
func ParseInt(str string) int {
	i, _ := strconv.Atoi(str)
	return i
}

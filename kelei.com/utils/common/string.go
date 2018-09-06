package common

import (
	"bytes"
)

//去掉最后一个多余的字符
func RemoveLastChar(buff bytes.Buffer) *string {
	str := buff.String()
	if str != "" {
		str = str[:len(str)-1]
	}
	return &str
}

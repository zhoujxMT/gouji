package common

import (
	"fmt"
	"strings"
)

type Address struct {
	IP   string
	Port string
}

func (a *Address) Remote() string {
	return fmt.Sprintf("%s:%d", a.IP, a.Port)
}

//根据配置文件加载地址
func NewAddress(addr string) *Address {
	addrInfo := strings.Split(addr, ":")
	ip, port := addrInfo[0], addrInfo[1]
	address := Address{ip, port}
	return &address
}

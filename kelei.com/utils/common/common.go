package common

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"kelei.com/utils/rpcs"
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

//获取两数之间随机数
func Random(min, max int) int {
	rand.Seed(time.Now().UnixNano())
	return min + rand.Intn(max-min)
}

//int数组的indexOf
func IndexIntOf(arr []int, val int) int {
	for i, v := range arr {
		if v == val {
			return i
		}
	}
	return -1
}

//map的indexOf
func IndexMapOf(m map[int]int, val int) int {
	for i, _ := range m {
		if i == val {
			return i
		}
	}
	return -1
}

//string数组的indexOf
func IndexStringOf(arr []string, val *string) int {
	for i, s := range arr {
		if s == *val {
			return i
		}
	}
	return -1
}

//将字符串数组转化成数字数组
func StrArrToIntArr(strArr []string) []int {
	intArr := make([]int, len(strArr))
	for i, str := range strArr {
		j, _ := strconv.Atoi(str)
		intArr[i] = j
	}
	return intArr
}

//将字符串转化成数字
func StrToInt(str string) int {
	i, _ := strconv.Atoi(str)
	return i
}

//将数字数组转化成字符串数组
func IntArrToStrArr(intArr []int) []string {
	strArr := make([]string, len(intArr))
	for i, v := range intArr {
		j := strconv.Itoa(v)
		strArr[i] = j
	}
	return strArr
}

//测试消耗时间
func CostTime(f func(), times int, title string) {
	t := time.Now().UnixNano()
	for i := 0; i < times; i++ {
		f()
	}
	totalCostTime := float32(time.Now().UnixNano()-t) / 1e6
	fmt.Println(title, "并发数:", times)
	fmt.Println(title, "消耗总时间(ms):", totalCostTime)
	avgTime := float32(totalCostTime) / float32(times)
	fmt.Println(title, "平均响应时间(ms):", avgTime)
	fmt.Println(title, "每秒钟吞吐量:", int(float32(times)/(float32(totalCostTime)/1000)))
}

//提取map[string]interface{}中的args转化成[]string
func ExtractArgs(rpcArgs *rpcs.Args) []string {
	args_ := rpcArgs.V["args"].([]interface{})
	args := []string{}
	for _, arg := range args_ {
		args = append(args, arg.(string))
	}
	return args
}

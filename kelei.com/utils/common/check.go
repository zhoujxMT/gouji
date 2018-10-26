package common

import (
	"strconv"
)

//检验参数是否是数字
func CheckNum(str string) (int, error) {
	num, err := strconv.Atoi(str)
	if err != nil {
		return 0, err
	}
	return num, nil
}

//检验参数是否是bool
func CheckBool(str string) (bool, error) {
	b, err := strconv.ParseBool(str)
	if err != nil {
		return false, err
	}
	return b, nil
}

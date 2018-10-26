package lua

import (
	io "io/ioutil"

	"kelei.com/utils/logger"
)

func LoadLuaFile(filename string) (string, error) {
	content, err := io.ReadFile(filename)
	if err != nil {
		logger.Errorf("加载lua文件失败", err)
		return "", err
	}
	return string(content), nil
}

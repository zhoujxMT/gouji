package json

import (
	json "encoding/json"
	io "io/ioutil"

	"kelei.com/utils/logger"
)

type JsonStruct struct {
}

func NewJsonStruct() *JsonStruct {
	return &JsonStruct{}
}

func (self *JsonStruct) Load(filename string, v interface{}) error {
	data, err := io.ReadFile(filename)
	if err != nil {
		logger.Errorf("加载文件失败:%s", err)
		return err
	}
	datajson := []byte(data)
	err = json.Unmarshal(datajson, v)
	if err != nil {
		logger.Errorf("反编译文件失败:%s", err)
		return err
	}
	return nil
}

func (self *JsonStruct) LoadByString(data string, v interface{}) error {
	datajson := []byte(data)
	err := json.Unmarshal(datajson, v)
	logger.CheckError(err, "json:LoadByString")
	return err
}

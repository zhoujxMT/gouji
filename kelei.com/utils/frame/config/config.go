package config

import (
	"kelei.com/utils/common"
	"kelei.com/utils/frame"
	"kelei.com/utils/logger"
)

func Mode(commVal string) {
	if commVal == "" {
		modeName := "未知模式"
		if frame.GetMode() == frame.MODE_RELEASE {
			modeName = "生产模式"
		} else if frame.GetMode() == frame.MODE_DEBUG {
			modeName = "调试模式"
		}
		logger.Infof(modeName)
	} else {
		if v, err := common.CheckNum(commVal); err == nil {
			frame.SetMode(v)
		}
	}
}

package logger

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestConfig(t *testing.T) {
	Convey("日志初始化", t, func() {
		So(Config("log/logger.log", 1), ShouldEqual, true)
	})
}

func TestCreateLogFile(t *testing.T) {
	Convey("创建日志文件", t, func() {
		So(createLogFile(), ShouldEqual, true)
	})
}

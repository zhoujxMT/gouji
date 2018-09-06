package cmds

import (
	eng "kelei.com/richman/engine"
	"kelei.com/utils/frame/command"
	"kelei.com/utils/logger"
)

var (
	engine *eng.Engine
)

func Inject(engine_ *eng.Engine) {
	engine = engine_
}

func GetCmds() map[string]func(string) {
	var commands = map[string]func(string){}
	//begin 自定义方法
	commands["test"] = test
	//end
	command.CreateHelp(commands)
	return commands
}

func test(cmdVal string) {
	logger.Infof("test")
}

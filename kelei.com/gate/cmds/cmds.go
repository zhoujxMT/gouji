package cmds

import (
	"kelei.com/utils/frame/command"
	"kelei.com/utils/frame/config"
)

func GetCmds() map[string]func(string) {
	var commands = map[string]func(string){}
	//begin 自定义方法

	//end
	commands["mode"] = config.Mode
	command.CreateHelp(commands)
	return commands
}

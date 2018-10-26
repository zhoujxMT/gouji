package cmds

import (
	"strings"

	"kelei.com/server/update"
	"kelei.com/utils/frame/command"
	"kelei.com/utils/frame/config"
)

func GetCmds() map[string]func(string) {
	var commands = map[string]func(string){}
	//begin 自定义方法
	commands["event"] = event
	//end
	commands["mode"] = config.Mode
	command.CreateHelp(commands)
	return commands
}
func event(commVal string) {
	if commVal == "" {
		return
	}
	vals := strings.Split(commVal, ",")
	name, triggerWeekday, triggerTime := vals[0], vals[1], vals[2]
	update.SetEvent(name, triggerWeekday, triggerTime)
}

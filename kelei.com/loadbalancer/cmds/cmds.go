package cmds

import (
	"fmt"

	"kelei.com/utils/frame/command"
)

func GetCmds() map[string]func(string) {
	var commands = map[string]func(string){}
	//begin 自定义方法
	commands["test"] = test
	//end
	command.CreateHelp(commands)
	return commands
}

func test(cmdVal string) {
	fmt.Println("good")
}

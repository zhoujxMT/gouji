package frame

import (
	"bufio"
	"os"

	"kelei.com/utils/frame/command"
	"kelei.com/utils/logger"
)

//日记模块
func loadLog() {
	logger.Config("log/.log", logger.DebugLevel)
}

//命令模块
func loadCommand() bool {
	commands := args.Commands
	var customCommandBlock func(string) bool
	if commands != nil {
		customCommandBlock = func(comm string) bool {
			success := false
			commName, commVal := command.ParseCommand(comm)
			for cmdName, cmd := range commands {
				if commName == cmdName {
					cmd(commVal)
					success = true
					break
				}
			}
			if !success {
				logger.Errorf("无效的命令 : %s %s", commName, commVal)
			}
			return success
		}
	} else {
		customCommandBlock = func(comm string) bool {
			logger.Debugf("没有自定义的命令块")
			return false
		}
	}
	command.Config()
	command.AppendCommandBlock(logger.CommandBlock())
	command.AppendCommandBlock(customCommandBlock)
	return true
}

//监听输入
func listenStdin() {
	input := bufio.NewScanner(os.Stdin)
	for input.Scan() {
		text := input.Text()
		command.AppendCommand(command.NewComm(text, false))
	}
}

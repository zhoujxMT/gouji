package command

import (
	"bytes"
	"fmt"
	"strings"
)

//将命令字符串解析成（命令、参数）
func ParseCommand(comm string) (string, string) {
	commInfo := strings.Split(comm, " ")
	commInfoCount := len(commInfo)
	commName, commVal := commInfo[0], ""
	if commInfoCount >= 2 {
		commVal = strings.Join(commInfo[1:], " ")
	}
	return commName, commVal
}

type Comm struct {
	content  string
	complete bool
}

func NewComm(content string, complete bool) *Comm {
	return &Comm{content, complete}
}
func (comm *Comm) GetContent() string {
	return comm.content
}

type Command struct {
	CommandList   []*Comm
	CommandBlocks []func(comm string) bool
}

var command Command

func Config() {
	command.CommandList = []*Comm{}
	command.CommandBlocks = []func(comm string) bool{}
	AppendCommandBlock(commandBlock)
}

func AppendCommand(comm *Comm) {
	command.CommandList = append(command.CommandList, comm)
	for _, commandBlock := range command.CommandBlocks {
		if commandBlock(comm.content) {
			break
		}
	}
}

func AppendCommandBlock(commandBlock func(comm string) bool) {
	command.CommandBlocks = append(command.CommandBlocks, commandBlock)
}

//根据输入执行对应命令
func commandBlock(comm string) bool {
	success := true
	commName, commVal := ParseCommand(comm)
	commVal = commVal
	switch commName {
	case "commandlist":
		for i, cmd := range command.CommandList {
			fmt.Printf("命令-%d : %s\n", i, cmd.GetContent())
		}
	default:
		success = false
	}
	return success
}

//生成帮助方法
func CreateHelp(commands map[string]func(string)) {
	h := func(cmdVal string) {
		buff := bytes.Buffer{}
		for name, _ := range commands {
			buff.WriteString(fmt.Sprintf("     %s\n", name))
		}
		info := buff.String()
		fmt.Printf("{\n%s}\n", info)
	}
	commands["h"] = h
}

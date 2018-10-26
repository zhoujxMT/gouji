package engine

import (
	"strconv"
	"strings"

	"kelei.com/utils/common"
	"kelei.com/utils/rpcs"
)

const (
	PushType_Notice = iota
	PushType_Email
)

//从server服务器推送过来的消息
func (this *Engine) PushMessage(args *rpcs.Args) *rpcs.Reply {
	p := this.getArgs(args)
	userids, funcName, messages := p[0], p[1], p[2]
	if userids == "*" {
		for _, user := range UserManage.GetAllUsers() {
			user.push(funcName, &messages)
		}
	} else {
		arrMessages := strings.Split(messages, "^")
		arrUserIds := strings.Split(userids, ",")
		this.handleRedisCache(arrMessages, arrUserIds)
		for i, userid := range arrUserIds {
			message := arrMessages[i]
			user := UserManage.GetUser(&userid)
			if user != nil {
				user.push(funcName, &message)
			}
		}
	}
	return &rpcs.Reply{nil, common.SC_OK}
}

//处理redis缓存
func (this *Engine) handleRedisCache(messages, userids []string) {
	for i, userid := range userids {
		user := UserManage.GetUser(&userid)
		if user != nil {
			info := strings.Split(messages[i], "|")
			type_, _ := strconv.Atoi(info[0])
			if type_ == PushType_Email {
				user.SetEmailExpire()
			}
		}
	}
}

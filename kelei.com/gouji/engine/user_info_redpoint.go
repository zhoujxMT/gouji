package engine

import (
	"strings"

	. "kelei.com/utils/common"
)

const (
	RedPointType_Honor = iota //成就
	RedPointType_Mail         //邮件
)

/*
获取红点信息根据类型
in:redPointType
out:0|1|0	(0没1有)
*/
func (u *User) GetRedPoint(redPointType int) *string {
	res := "1"
	if redPointType == RedPointType_Mail {
		res = u.GetMailRP()
	} else if redPointType == RedPointType_Honor {
		res = u.GetHonorRP()
	}
	return &res
}

//获取邮件红点
func (u *User) GetMailRP() string {
	res := "0"
	mails_ := *u.GetMails()
	if mails_ != Res_NoData {
		mails := strings.Split(mails_, "#")
		var recestatus, isread string
		for _, mail_ := range mails {
			mail := strings.Split(mail_, "@")
			recestatus, isread = mail[4], mail[6]
			//未领取奖励 || 未读
			if recestatus == "0" || isread == "0" {
				res = "1"
				break
			}
		}
	}
	return res
}

//获取成就红点
func (u *User) GetHonorRP() string {
	res := ""
	honorTypes := []int{1, 2, 3, 4, 5, 6}
	redPoint := []string{"0", "0", "0", "0", "0", "0"}
	for i, honorType := range honorTypes {
		honorInfo := *u.GetHonorInfo(honorType)
		if honorInfo == "-1" {
			continue
		}
		honorInfoArr := strings.Split(honorInfo, "|")
		for _, honor_s := range honorInfoArr {
			honor := strings.Split(honor_s, "$")
			if honor[3] == "1" {
				redPoint[i] = "1"
				break
			}
		}
	}
	res = strings.Join(redPoint, "|")
	return res
}

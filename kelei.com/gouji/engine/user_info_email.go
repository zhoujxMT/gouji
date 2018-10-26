package engine

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/garyburd/redigo/redis"

	. "kelei.com/utils/common"
	"kelei.com/utils/logger"
)

func (u *User) getEmailKey() string {
	return fmt.Sprintf("email:%s", *u.getUserID())
}

//加载邮件表
func (u *User) loadMail() *string {
	gameRds := getGameRds()
	defer gameRds.Close()
	key := u.getEmailKey()
	info, err := redis.String(gameRds.Do("get", key))
	if err != nil {

		rows, err := db.Query("call MailGet(?);", *u.getUserID())
		logger.CheckFatal(err, "GetMails")
		defer rows.Close()
		buff := bytes.Buffer{}
		id, senduserid, recestatus := 0, 0, 0
		isread := []uint8{}
		title, content, prize, createtime := "", "", "", ""
		for rows.Next() {
			rows.Scan(&id, &senduserid, &title, &content, &prize, &recestatus, &createtime, &isread)
			buff.WriteString(fmt.Sprintf("%d@%d@%s@%s@%s@%d@%s@%d#", id, senduserid, title, content, prize, recestatus, createtime, int(isread[0])))
		}
		info = *RowsBufferToString(buff)
		gameRds.Do("set", key, info)
		gameRds.Do("expire", key, expireSecond_Long)
	}
	return &info
}

/*
获取邮件列表
in:
out:-1没有邮件
	id@发送人@标题@领取状态@创建时间@读取标识#
des:发送人 0系统 >0玩家
	奖励 itemid|count$
	领取状态 0未领取 1已领取 2不需要领取
	读取标识 0未读取 1已读取
*/
func (u *User) GetMails() *string {
	mails_s := u.loadMail()
	//没有邮件
	if *mails_s == Res_NoData {
		return mails_s
	}
	mails := strings.Split(*mails_s, "#")
	buff := bytes.Buffer{}
	var id, senduserid, recestatus, isread, title, createtime string
	for _, mail_s := range mails {
		mail := strings.Split(mail_s, "@")
		id, senduserid, title, recestatus, createtime, isread = mail[0], mail[1], mail[2], mail[5], mail[6], mail[7]
		buff.WriteString(fmt.Sprintf("%s@%s@%s@@%s@%s@%s#", id, senduserid, title, recestatus, createtime, isread))
	}
	info := *RowsBufferToString(buff)
	return &info
}

/*
读取邮件
in:mailid
out:-1邮件不存在
	id@发件人@标题@内容@奖励@领取状态@创建时间@读取标识
*/
func (u *User) ReadMail(mailid int) *string {
	res := Res_NoData
	//获取邮件是否存在
	res = *u.getMailInfo(mailid)
	//邮件不存在
	if res == Res_NoData {
		return &res
	}
	//检测邮件是否已读
	mail := strings.Split(res, "@")
	isRead := mail[7]
	//已读
	if isRead == "1" {
		return &res
	}
	//未读更新数据库

	_, err := db.Exec("update Email set IsRead=1 where Email.ID=? and Email.ReceUserID=?;", mailid, *u.getUserID())
	logger.CheckFatal(err, "ReadMail")
	u.SetEmailExpire()
	res = *u.getMailInfo(mailid)
	return &res
}

/*
获取单个邮件信息
*/
func (u *User) getMailInfo(mailid int) *string {
	mails_s := u.loadMail()
	if *mails_s == Res_NoData {
		return mails_s
	}
	res := Res_NoData
	mails := strings.Split(*mails_s, "#")
	for _, mail_s := range mails {
		mail := strings.Split(mail_s, "@")
		id := mail[0]
		if strconv.Itoa(mailid) == id {
			res = mail_s
			break
		}
	}
	return &res
}

/*
领取邮件奖励
in:mailid(-1领取全部邮件奖励 >0领取单个邮件领取)
out:-1邮件不存在 -2没有奖励可领取 -3奖励已领取
	itemid|count$
*/
func (u *User) ReceMail(mailid int) *string {
	res := &Res_NoData
	if mailid == -1 {
		res = u.ReceAllMail()
	} else {
		res = u.ReceOneMail(mailid)
	}
	return res
}

/*
ReceMail（领取单个邮件奖励）
in:mailid(-1一键领取 >0单个领取)
out:-1邮件不存在 -2没有奖励可领取 -3奖励已领取
	itemid|count$
*/
func (u *User) ReceOneMail(mailid int) *string {
	res := Res_NoData
	//获取邮件是否存在
	res = *u.getMailInfo(mailid)
	//邮件不存在
	if res == Res_NoData {
		return &res
	}
	//检测邮件奖励领取状态
	mail := strings.Split(res, "@")
	recestatus, _ := strconv.Atoi(mail[5])
	//没有奖励可领取
	if recestatus == 2 {
		res = "-2"
		return &res
	}
	//奖励已领取
	if recestatus == 1 {
		res = "-3"
		return &res
	}
	//领取奖励

	err := db.QueryRow("call MailRece(?,?,@res);", *u.getUserID(), mailid).Scan(&res)
	logger.CheckFatal(err, "ReceOneMail")
	u.SetEmailExpire()
	return &res
}

/*
领取所有邮件奖励
in:mailid()
out:-1邮件不存在 -2没有奖励可领取
	itemid|count$
*/
func (u *User) ReceAllMail() *string {
	res := Res_NoData
	mails_s := u.loadMail()
	//没有邮件
	if *mails_s == Res_NoData {
		return mails_s
	}
	//有没有可领取奖励的邮件
	haveCanRece := false
	mails := strings.Split(*mails_s, "#")
	//检测所有邮件奖励领取状态
	for _, mail_s := range mails {
		mail := strings.Split(mail_s, "@")
		recestatus := mail[5]
		if recestatus == "0" {
			haveCanRece = true
			break
		}
	}
	//没有奖励可领取
	if !haveCanRece {
		res = "-2"
		return &res
	}
	//领取全部奖励,存储过程中没有添加道具,返回重复的道具字符串,放到下面合并一下再给玩家添加

	err := db.QueryRow("call MailAllRece(?);", *u.getUserID()).Scan(&res)
	logger.CheckFatal(err, "ReceAllMail")
	if res == "" {
		res = "-2"
	}
	u.SetEmailExpire()
	return &res
}

//合并相同道具
func (u *User) mergeItems(items_s string) *string {
	mapItemInfo := map[string]int{}
	items := strings.Split(items_s, "$")
	for _, item_s := range items {
		item := strings.Split(item_s, "|")
		itemid, itemcount_s := item[0], item[1]
		itemcount, err := strconv.Atoi(itemcount_s)
		logger.CheckFatal(err)
		mapItemInfo[itemid] = mapItemInfo[itemid] + itemcount
	}
	itemsinfo := bytes.Buffer{}
	for itemid, itemcount := range mapItemInfo {
		itemsinfo.WriteString(fmt.Sprintf("%s|%d$", itemid, itemcount))
	}
	res := itemsinfo.String()
	res = res[:len(res)-1]
	return &res
}

//设置用户的邮件过期
func (u *User) SetEmailExpire() {
	gameRds := getGameRds()
	defer gameRds.Close()
	gameRds.Do("expire", u.getEmailKey(), 0)
}

package common

import (
	"gopkg.in/gomail.v2"

	"kelei.com/utils/logger"
)

/*
发送邮件
in:主题,内容
*/
func SendMail(subject, content string) {
	m := gomail.NewMessage()
	m.SetHeader("From", "399751568@qq.com")
	m.SetHeader("To", "904069295@qq.com")
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", content)

	d := gomail.NewDialer("smtp.qq.com", 465, "399751568@qq.com", "otifkonkmvuvbicc")

	// Send the email to Bob, Cora and Dan.
	err := d.DialAndSend(m)
	logger.CheckError(err)
}

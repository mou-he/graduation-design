package email

import (
	"fmt"

	"github.com/mou-he/graduation-design/config"
	"gopkg.in/gomail.v2"
)

const (
	CodeMsg     = "HaiAI验证码如下(验证码仅限3分钟有效)"
	UserNameMsg = "GopherAI的账号如下，请保留好，后续可以用账号/邮箱登录 "
)

func SendCaptcha(email, code, msg string) error {

	// 发送邮件
	m := gomail.NewMessage()
	m.SetHeader("From", config.GetConfig().Email)
	m.SetHeader("To", email)
	m.SetHeader("Subject", "HaiAI验证码")
	m.SetBody("text/html", fmt.Sprintf("%s %s", msg, code))
	// 配置 SMTP 服务器和授权码,587：是 SMTP 的明文/STARTTLS 端口号
	d := gomail.NewDialer("smtp.qq.com", 587, config.GetConfig().EmailConfig.Email, config.GetConfig().EmailConfig.Authcode)

	// 发送邮件
	if err := d.DialAndSend(m); err != nil {
		fmt.Printf("DialAndSend err %v:\n", err)
		return err
	}
	fmt.Printf("send mail success\n")
	return nil

}

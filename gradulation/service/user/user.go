package user

import (
	"github.com/mou-he/graduation-design/common/code"
	myemail "github.com/mou-he/graduation-design/common/email"
	"github.com/mou-he/graduation-design/common/redis"
	"github.com/mou-he/graduation-design/dao/user"
	dao "github.com/mou-he/graduation-design/dao/user"
	"github.com/mou-he/graduation-design/model"
	"github.com/mou-he/graduation-design/utils"
	"github.com/mou-he/graduation-design/utils/myjwt"
)

func Login(username, password string) (string, code.Code) {
	var userInformation *model.User
	var ok bool
	//1:判断用户是否存在
	if ok, userInformation = dao.IsUserExist(username); !ok {

		return "", code.CodeUserNotExist
	}
	//2:判断用户是否密码账号正确
	if userInformation.Password != utils.MD5(password) {
		return "", code.CodeInvalidPassword
	}
	//3:返回一个Token
	token, err := myjwt.GenerateToken(userInformation.ID, userInformation.Username)

	if err != nil {
		return "", code.CodeServerBusy
	}
	return token, code.CodeSuccess
}

func Register(email, password, captcha string) (string, code.Code) {
	// 检查用户是否存在
	var userInfo *model.User
	// 判断用户是否已存在
	if ok, _ := user.IsUserExist(email); ok {
		return "", code.CodeUserExist
	}
	// 验证验证码
	if ok, _ := redis.CheckCaptchaForEmail(email, captcha); !ok {
		return "", code.CodeInvalidCaptcha
	}
	// 创建用户账号
	username := utils.GetRandomNumbers(11)
	// 注册到数据库中
	userInfo, ok := user.Register(username, email, password)
	if !ok {
		return "", code.CodeServerBusy
	}
	// 发送账号到用户邮箱
	if err := myemail.SendCaptcha(email, username, myemail.UserNameMsg); err != nil {
		return "", code.CodeServerBusy
	}
	// 生成token
	token, err := myjwt.GenerateToken(userInfo.ID, userInfo.Username)
	if err != nil {
		return "", code.CodeServerBusy
	}
	// 返回账号
	return token, code.CodeSuccess
}

// 发送验证码的函数
func SendCaptcha(email string) code.Code {
	send_code := utils.GetRandomNumbers(6)
	// 发送验证码
	if err := redis.SetCaptchaForEmail(email, send_code); err != nil {
		return code.CodeServerBusy
	}

	// 发送验证码
	if err := myemail.SendCaptcha(email, send_code, myemail.CodeMsg); err != nil {
		return code.CodeServerBusy
	}
	return code.CodeSuccess
}

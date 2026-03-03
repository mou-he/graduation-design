package user

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/mou-he/graduation-design/common/code"
	myemail "github.com/mou-he/graduation-design/common/email"
	"github.com/mou-he/graduation-design/common/redis"
	"github.com/mou-he/graduation-design/dao/user"
	dao "github.com/mou-he/graduation-design/dao/user"
	"github.com/mou-he/graduation-design/model"
	"github.com/mou-he/graduation-design/utils"
	"github.com/mou-he/graduation-design/utils/myjwt"
)

func Login(c *gin.Context, username, password string) (string, code.Code) {
	var userInformation *model.User
	var existUser bool
	var existEmail bool
	//1:判断用户是否存在
	if existUser, userInformation = dao.IsUserExist(username); !existUser {
		if existEmail, userInformation = dao.IsEmailExist(username); !existEmail {
			fmt.Println(userInformation)
			if userInformation == nil {
				fmt.Println(userInformation)
				return "", code.CodeUserNotExist
			}
		}
	}
	fmt.Printf("存储密码 %v\n", userInformation.Password)
	fmt.Printf("进入密码 %s\n", utils.MD5(password))
	//2:判断用户是否密码账号正确
	if userInformation.Password != utils.MD5(password) {
		return "", code.CodeInvalidPassword
	}
	//3:返回一个Token
	token, err := myjwt.GenerateToken(userInformation.ID, userInformation.Username)
	c.Set("username", userInformation.Username)
	c.Set("user_id", userInformation.ID)
	if err != nil {
		return "", code.CodeServerBusy
	}
	return token, code.CodeSuccess
}

func Register(c *gin.Context, email, password, captcha string) (string, code.Code) {
	// 检查用户是否存在
	var userInfo *model.User
	// 判断用户是否已存在
	if ok, _ := user.IsEmailExist(email); ok {
		return "", code.CodeEmailExist
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
	c.Set("username", username)
	c.Set("user_id", userInfo.ID)
	// 返回账号
	return token, code.CodeSuccess
}

// 发送注册验证码的函数
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

func SendForgetPasswordCaptcha(email string) code.Code {
	// 1. 检查邮箱是否存在
	if ok, _ := dao.IsEmailExist(email); !ok {
		return code.CodeEmailNotExist // 邮箱不存在，无法找回密码
	}

	// 2. 生成6位随机数字验证码
	send_code := utils.GetRandomNumbers(6)

	fmt.Printf("Generated forget password captcha for %s: %s\n", email, send_code) // 打印验证码方便调试

	// 3. 将验证码存储到 Redis，设置过期时间 (例如 5 分钟)
	if err := redis.SetCaptchaForEmail(email, send_code); err != nil {
		fmt.Printf("Failed to set captcha for email %s in redis: %v\n", email, err)
		return code.CodeServerBusy
	}

	// 4. 发送验证码到用户邮箱 (使用 myemail.ForgetPasswordCodeMsg 模板)
	if err := myemail.SendCaptcha(email, send_code, myemail.ForgetPasswordCodeMsg); err != nil {
		fmt.Printf("Failed to send forget password captcha to email %s: %v\n", email, err)
		return code.CodeServerBusy
	}
	return code.CodeSuccess
}

func ResetPassword(email, captcha, newPassword string) code.Code {
	// 1. 检查邮箱是否存在
	user, err := dao.GetUserByEmail(email)
	if err != nil || user == nil {
		return code.CodeEmailNotExist
	}

	// 2. 验证验证码
	if ok, _ := redis.CheckCaptchaForEmail(email, captcha); !ok {
		return code.CodeInvalidCaptcha
	}
	fmt.Printf("password %s\n", newPassword)
	// 3. 更新用户密码
	hashedPassword := utils.MD5(newPassword)
	if err := dao.UpdateUserPassword(user.ID, hashedPassword); err != nil {
		fmt.Printf("Failed to update password for user %s: %v\n", email, err)
		return code.CodeServerBusy
	}
	return code.CodeSuccess
}

package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	mycode "github.com/mou-he/graduation-design/common/code"
	controller "github.com/mou-he/graduation-design/controller"
	userservice "github.com/mou-he/graduation-design/service/user"
)

type (
	LoginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	RegisterRequset struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Captcha  string `json:"captcha"`
	}
	RegisterResponse struct {
		controller.Response
		Token string `json:"token"`
	}
	LoginResponse struct {
		controller.Response
		Token string `json:"token"`
	}
	CaptchaRequest struct {
		Email string `json:"email"`
	}
	CaptchaResponse struct {
		controller.Response
	}
	ForgetPasswordCaptchaRequest struct {
		Email string `json:"email" binding:"required,email"`
	}
	ForgetPasswordCaptchaResponse struct {
		controller.Response
	}

	ResetPasswordRequest struct {
		Email           string `json:"email" binding:"required,email"`
		Captcha         string `json:"captcha" binding:"required,len=6"`
		NewPassword     string `json:"new_password" binding:"required,min=6,max=20"`
		ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=NewPassword"`
	}
	ResetPasswordResponse struct {
		controller.Response
	}
)

func Register(c *gin.Context) {

	var req RegisterRequset
	var res RegisterResponse
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, res.CodeOf(mycode.CodeInvalidParams))
		return
	}
	// 注册
	token, code := userservice.Register(c, req.Email, req.Password, req.Captcha)
	if code != mycode.CodeSuccess {
		c.JSON(http.StatusOK, res.CodeOf(code))
		return
	}
	// 注册成功
	res.Success()
	// 返回token
	res.Token = token
	c.JSON(http.StatusOK, res)

}

func Login(c *gin.Context) {
	var req LoginRequest
	var res LoginResponse
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, res.CodeOf(mycode.CodeInvalidParams))
		return
	}
	// 登录
	token, code := userservice.Login(c, req.Username, req.Password)
	if code != mycode.CodeSuccess {
		c.JSON(http.StatusOK, res.CodeOf(code))
		return
	}

	// 登录成功
	res.Success()
	// 返回token
	res.Token = token
	c.JSON(http.StatusOK, res)

}
func Captcha(c *gin.Context) {
	{
		var req CaptchaRequest
		var res CaptchaResponse
		if err := c.ShouldBind(&req); err != nil {
			c.JSON(http.StatusBadRequest, res.CodeOf(mycode.CodeInvalidParams))
			return
		}
		code := userservice.SendCaptcha(req.Email)
		if code != mycode.CodeSuccess {
			c.JSON(http.StatusOK, res.CodeOf(code))
			return
		}
		// 验证码发送成功
		res.Success()
		// 返回验证码
		c.JSON(http.StatusOK, res)

	}
}
func ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	var res ResetPasswordResponse
	if err := c.ShouldBind(&req); err != nil {
		// elog.Error("参数校验失败", zap.Error(err)) // 非必要，根据你的日志框架添加
		c.JSON(http.StatusBadRequest, res.CodeOf(mycode.CodeInvalidParams))
		return
	}

	// 这里服务层只需要 NewPassword，ConfirmPassword 只是前端和请求层校验用
	code := userservice.ResetPassword(req.Email, req.Captcha, req.NewPassword)
	if code != mycode.CodeSuccess {
		c.JSON(http.StatusOK, res.CodeOf(code))
		return
	}

	res.Success()
	c.JSON(http.StatusOK, res)
}

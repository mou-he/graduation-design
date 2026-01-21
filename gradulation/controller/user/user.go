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
)

func Register(c *gin.Context) {

	var req RegisterRequset
	var res RegisterResponse
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, res.CodeOf(mycode.CodeInvalidParams))
		return
	}
	// 注册
	token, code := userservice.Register(req.Email, req.Password, req.Captcha)
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
	token, code := userservice.Login(req.Username, req.Password)
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

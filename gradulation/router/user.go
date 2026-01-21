package router

import (
	"github.com/gin-gonic/gin"
	usercontroller "github.com/mou-he/graduation-design/controller/user"
)

func RegisterUserRouter(r *gin.RouterGroup) {
	{
		r.POST("/register", usercontroller.Register)
		r.POST("/login", usercontroller.Login)
		r.POST("/captcha", usercontroller.Captcha)
	}
}

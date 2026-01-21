package router

import (
	"github.com/gin-gonic/gin"
	"github.com/mou-he/graduation-design/middleware/jwt"
)

func InitRouter() *gin.Engine {
	r := gin.Default()
	enterRouter := r.Group("/api/v1")
	{
		RegisterUserRouter(enterRouter.Group("/user"))
	}
	enterRouter.Use(jwt.Auth())
	{
		AIGroup := enterRouter.Group("/AI")
		AIGroup.Use(jwt.Auth())
		AIRouter(AIGroup)
	}
	return r
}

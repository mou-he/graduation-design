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
	{
		AIGroup := enterRouter.Group("/AI")
		AIGroup.Use(jwt.Auth())
		AIRouter(AIGroup)
	}
	{
		ImageGroup := enterRouter.Group("/image")
		ImageGroup.Use(jwt.Auth())
		ImageRouter(ImageGroup)
	}
	{
		FileGroup := enterRouter.Group("/file")
		FileGroup.Use(jwt.Auth())
		FileRouter(FileGroup)
	}
	{
		VideoGroup := enterRouter.Group("/video")
		VideoGroup.Use(jwt.Auth()) // 开启鉴权
		VideoRouter(VideoGroup)
	}
	{
		SuggestionGroup := enterRouter.Group("/suggestion")
		SuggestionGroup.Use(jwt.Auth())
		SuggestionRouter(SuggestionGroup)
	}

	return r
}

package router

import (
	"github.com/gin-gonic/gin"
	video "github.com/mou-he/graduation-design/controller/video" // 确保引用了你的controller包
)

func VideoRouter(r *gin.RouterGroup) {
	r.POST("/generate", video.GenerateVideo)
	r.GET("/status/:id", video.GetVideoStatus)
	r.GET("/play/:id", video.ProxyVideoPlay)
}

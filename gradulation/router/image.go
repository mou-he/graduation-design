package router

import (
	"github.com/gin-gonic/gin"
	"github.com/mou-he/graduation-design/controller/image"
)

func ImageRouter(r *gin.RouterGroup) {
	r.POST("/image/recognize", image.RecognizeImage)
}

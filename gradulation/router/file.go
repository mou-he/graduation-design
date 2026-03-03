package router

import (
	"github.com/gin-gonic/gin"
	"github.com/mou-he/graduation-design/controller/file"
)

func FileRouter(r *gin.RouterGroup) {
	r.POST("/upload", file.UploadRagFile)
}

package file

import (
	"log"
	"net/http"

	gin "github.com/gin-gonic/gin"
	"github.com/mou-he/graduation-design/common/code"
	"github.com/mou-he/graduation-design/controller"
	"github.com/mou-he/graduation-design/service/file"
)

type UploadFileResponse struct {
	FilePath string `json:"file_path,omitempty"`
	controller.Response
}

func UploadRagFile(c *gin.Context) {
	res := new(UploadFileResponse)
	uploadFile, err := c.FormFile("file")
	if err != nil {
		log.Println("FormFile fail ", err)
		c.JSON(http.StatusOK, res.CodeOf(code.CodeInvalidParams))
		return
	}
	username := c.GetString("username")
	// 从请求中获取文件
	if username == "" {
		log.Println("Username not found in context")
		c.JSON(http.StatusOK, res.CodeOf(code.CodeInvalidToken))
		return
	}
	// indexer 创建
	filePath, err := file.UploadRagFile(username, uploadFile)
	if err != nil {
		log.Printf("Error uploading file: %v", err)
		c.JSON(http.StatusOK, res.CodeOf(code.CodeServerBusy))
		return
	}
	res.Success()
	// 响应
	res.FilePath = filePath
	c.JSON(http.StatusOK, res)
}

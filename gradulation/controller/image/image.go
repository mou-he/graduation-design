package image

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mou-he/graduation-design/common/code"
	"github.com/mou-he/graduation-design/controller"
	"github.com/mou-he/graduation-design/service/image"
)

type (
	RecognizeImageRes struct {
		ClassName string `json:"className"`
		controller.Response
	}
)

func RecognizeImage(c *gin.Context) {
	res := new(RecognizeImageRes)
	file, err := c.FormFile("image")
	if err != nil {
		log.Println("file fail err is : ", err)
		c.JSON(http.StatusOK, res.CodeOf(code.CodeInvalidParams))
		return
	}
	className, err := image.RecognizeImage(file)
	if err != nil {
		log.Println("RecognizeImage fail : ", err)
		c.JSON(http.StatusOK, res.CodeOf(code.CodeServerBusy))
		return
	}
	res.Success()
	res.ClassName = className
	c.JSON(http.StatusOK, res)
}

package jwt

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mou-he/graduation-design/common/code"
	"github.com/mou-he/graduation-design/controller"
	"github.com/mou-he/graduation-design/utils/myjwt"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		res := new(controller.Response)
		var token string
		auth := c.GetHeader("Authorization")
		if auth != "" && strings.HasPrefix(auth, "Bearer ") {
			token = strings.TrimPrefix(auth, "Bearer ")
		}
		if token == "" {
			fmt.Println("Authorization header is empty")
			c.JSON(http.StatusOK, res.CodeOf(code.CodeInvalidToken))
			c.Abort()
			return
		}
		log.Println("token is", token)
		userName, ok := myjwt.ParseToken(token)
		if !ok {
			fmt.Println("ParseToken failed")
			c.JSON(http.StatusOK, res.CodeOf(code.CodeInvalidToken))
			c.Abort()
			return
		}
		c.Set("username", userName)
		c.Next()
	}
}

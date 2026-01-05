package main

import (
	// "fmt"

	"github.com/gin-gonic/gin"
	// "github.com/mou-he/graduation-design/router"
)

func StartServer(addr string, port int) error {
	// r := router.InitRouter()
	// return r.Run(fmt.Sprintf("%s:%d", addr, port))
	return nil
}

func main() {
	router := gin.Default()
	router.GET("/", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"message": "hello world",
		})
	})
	router.Run(":9090")
}

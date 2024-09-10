package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	r := InitWebServer()
	r.GET("/hello", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "hello")
	})
	r.Run(":8080")
}

package web

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"testing"
)

func TestNil(t *testing.T) {

}

func TestGin(t *testing.T) {
	r := gin.Default()
	r.GET("/hello", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "hello")
	})
	r.Run("0.0.0.0:8080")
}

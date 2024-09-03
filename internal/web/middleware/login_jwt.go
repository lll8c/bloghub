package middleware

import (
	"encoding/gob"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"strings"
	"time"
)

// LoginJWTMiddlewareBuilder JWT登录校验
type LoginJWTMiddlewareBuilder struct {
	paths []string
}

func NewLoginJWTMiddlewareBuilder() *LoginJWTMiddlewareBuilder {
	return &LoginJWTMiddlewareBuilder{}
}

func (l *LoginJWTMiddlewareBuilder) IgnorePaths(path string) *LoginJWTMiddlewareBuilder {
	l.paths = append(l.paths, path)
	return l
}

// Build 登录校验或者刷新登录状态
func (l *LoginJWTMiddlewareBuilder) Build() gin.HandlerFunc {
	//用go的方式编码解码
	gob.Register(time.Now())
	return func(ctx *gin.Context) {
		//登录和注册不需要校验
		for _, path := range l.paths {
			if ctx.Request.URL.Path == path {
				return
			}
		}
		//Bearer xxx.xxx.xxx
		tokenHeader := ctx.GetHeader("authorization")
		if tokenHeader == "" {
			//没登录
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		segs := strings.Split(tokenHeader, " ")
		if len(segs) != 2 {
			//格式不对，没登录
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		tokenStr := segs[1]
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return []byte("secret"), nil
		})
		if err != nil {
			//没登录
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		//token无效
		if token == nil || !token.Valid {
			//没登录
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
	}
}

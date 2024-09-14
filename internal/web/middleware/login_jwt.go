package middleware

import (
	"encoding/gob"
	jwt2 "geektime/webook/internal/web/jwt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"time"
)

// LoginJWTMiddlewareBuilder JWT登录校验
type LoginJWTMiddlewareBuilder struct {
	jwt2.JwtHandler
	paths []string
}

func NewLoginJWTMiddlewareBuilder(jwtHdl jwt2.JwtHandler) *LoginJWTMiddlewareBuilder {
	return &LoginJWTMiddlewareBuilder{
		JwtHandler: jwtHdl,
	}
}

func (l *LoginJWTMiddlewareBuilder) IgnorePaths(path string) *LoginJWTMiddlewareBuilder {
	l.paths = append(l.paths, path)
	return l
}

// Build 登录校验
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
		//提取token

		tokenStr := l.ExtractToken(ctx)
		//解析时传指针
		claims := &jwt2.UserClaims{}
		//解析token
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte("secret"), nil
		})
		if err != nil {
			//没登录
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		//token无效
		if token == nil || !token.Valid || claims.Uid == 0 {
			//没登录
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		//增强登录安全，校验UserAgent
		if claims.UserAgent != ctx.Request.UserAgent() {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		//检测用户是否已经退出登录
		err = l.CheckSession(ctx, claims.Ssid)
		if err != nil {
			//已经退出登录
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		ctx.Set("claims", claims)
	}
}

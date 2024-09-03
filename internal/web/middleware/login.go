package middleware

import (
	"encoding/gob"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type LoginMiddlewareBuilder struct {
	paths []string
}

func NewLoginMiddlewareBuilder() *LoginMiddlewareBuilder {
	return &LoginMiddlewareBuilder{}
}

func (l *LoginMiddlewareBuilder) IgnorePaths(path string) *LoginMiddlewareBuilder {
	l.paths = append(l.paths, path)
	return l
}

// Build 登录校验或者刷新登录状态
func (l *LoginMiddlewareBuilder) Build() gin.HandlerFunc {
	//用go的方式编码解码
	gob.Register(time.Now())
	return func(ctx *gin.Context) {
		//登录和注册不需要校验
		for _, path := range l.paths {
			if ctx.Request.URL.Path == path {
				return
			}
		}
		sess := sessions.Default(ctx)
		if sess == nil {
			//没有登录
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		id := sess.Get("userId")
		if id == nil {
			//没有登录
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		updateTime := sess.Get("update_time")
		//当前时间time.Time
		now := time.Now()
		//还没有刷新过刚登录
		if updateTime == nil {
			sess.Set("update_time", now)
			sess.Save()
			return
		}
		updateTimeVal, _ := updateTime.(time.Time)
		if now.Sub(updateTimeVal) > time.Second*10 {
			sess.Set("update_time", now)
			sess.Save()
			return
		}
	}
}

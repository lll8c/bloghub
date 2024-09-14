package ioc

import (
	"context"
	"geektime/webook/internal/web"
	jwt2 "geektime/webook/internal/web/jwt"
	"geektime/webook/internal/web/middleware"
	"geektime/webook/pkg/logger"
	"github.com/fsnotify/fsnotify"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"strings"
	"time"
)

func InitWebServer(mdls []gin.HandlerFunc,
	userHandler *web.UserHandler, wechatHandler *web.OAuth2WechatHandler) *gin.Engine {
	r := gin.Default()
	r.Use(mdls...)
	//注册用户路由
	userHandler.RegisterRoutes(r)
	wechatHandler.RegisterRoutes(r)
	return r
}

func InitMiddlewares(jwtHdl jwt2.JwtHandler, l logger.LoggerV1) []gin.HandlerFunc {
	//web请求日志打印 配置
	bd := middleware.NewLogMiddlewareBuilder(func(ctx context.Context, al *middleware.AccessLog) {
		l.Debug("HTTP请求", logger.Field{Key: "al", Val: al})
	}).AllowReqBody(true).AllowRespBody(true)
	viper.OnConfigChange(func(in fsnotify.Event) {
		ok := viper.GetBool("web.logreq")
		bd.AllowReqBody(ok)
	})

	return []gin.HandlerFunc{
		//解决跨域问题
		cors.New(cors.Config{
			//允许跨域访问的客户端地址，*表示所有都允许
			//AllowOrigins: []string{"*"},
			//允许跨域访问的方法，默认都有
			//AllowMethods:     []string{"PUT", "PATCH"},
			//允许客户端使用的请求头
			AllowHeaders: []string{"Content-type", "Authorization"},
			//允许带cookie之类的用户认证信息
			AllowCredentials: true,
			//允许包含的请求
			ExposeHeaders: []string{"x-jwt-token", "x-refresh-token"},
			//使用方法校验忽略掉AllowOrigins
			AllowOriginFunc: func(origin string) bool {
				if strings.Contains(origin, "localhost") {
					return true
				}
				return strings.Contains(origin, "yourcompany.com")
			},
			//你的开发环境，包含localhost就允许访问
			MaxAge: 12 * time.Hour,
		}),
		//日志打印请求和响应
		bd.Build(),
		//JWT中间件校验
		//在校验中忽略某些路由
		middleware.NewLoginJWTMiddlewareBuilder(jwtHdl).
			IgnorePaths("/users/signup").
			IgnorePaths("/users/login").
			IgnorePaths("/hello").
			IgnorePaths("/users/login_sms/code/send").
			IgnorePaths("/login_sms").
			IgnorePaths("/oauth2/wechat/authurl").
			IgnorePaths("/oauth2/wechat/callback").
			IgnorePaths("/refresh_token").
			Build(),
	}
}

package ioc

import (
	"geektime/webook/internal/web"
	"geektime/webook/internal/web/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"strings"
	"time"
)

func InitWebServer(mdls []gin.HandlerFunc, userHandler *web.UserHandler) *gin.Engine {
	r := gin.Default()
	r.Use(mdls...)
	//注册用户路由
	userHandler.RegisterRoutes(r)
	return r
}

func InitMiddlewares() []gin.HandlerFunc {
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
			ExposeHeaders: []string{"x-jwt-token"},
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
		//JWT中间件校验
		//在校验中忽略某些路由
		middleware.NewLoginJWTMiddlewareBuilder().
			IgnorePaths("/users/signup").
			IgnorePaths("/users/login").
			IgnorePaths("/hello").
			IgnorePaths("/users/login_sms/code/send").
			IgnorePaths("/login_sms").
			Build(),
	}
}

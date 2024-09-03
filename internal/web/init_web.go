package web

import (
	"geektime/webook/internal/web/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"strings"
	"time"
)

func RegisterRoutes() *gin.Engine {
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		//解决跨域问题
		//允许跨域访问的客户端地址，*表示所有都允许
		//AllowOrigins: []string{"*"},
		//允许跨域访问的方法，默认都有
		//AllowMethods:     []string{"PUT", "PATCH"},
		//允许客户端使用的请求头
		AllowHeaders: []string{"Content-type", "Authorization"},
		//允许带cookie之类的用户认证信息
		AllowCredentials: true,
		//使用方法校验忽略掉AllowOrigins
		AllowOriginFunc: func(origin string) bool {
			if strings.Contains(origin, "localhost") {
				return true
			}
			return strings.Contains(origin, "yourcompany.com")
		},
		//你的开发环境，包含localhost就允许访问
		MaxAge: 12 * time.Hour,
	}))

	//使用session校验用户是否登录
	//创建基于 cookie 的存储引擎，secret 参数是用于加密的密钥
	//store := cookie.NewStore([]byte("secret"))
	//创建基于 redis 的存储引擎
	store, err := redis.NewStore(16, "tcp", "localhost:6379", "", []byte("secret"))
	if err != nil {
		panic(err)
	}
	//在ctx中创建session
	r.Use(sessions.Sessions("mysession", store))
	//校验
	r.Use(middleware.NewLoginMiddlewareBuilder().IgnorePaths("/users/signup").
		IgnorePaths("/users/login").Build())
	return r
}

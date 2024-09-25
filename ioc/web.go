package ioc

import (
	"geektime/webook/internal/web"
	jwt2 "geektime/webook/internal/web/jwt"
	"geektime/webook/internal/web/middleware"
	"geektime/webook/pkg/ginx/middlewares/metric"
	"geektime/webook/pkg/logger"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"strings"
	"time"
)

func InitWebServer(mdls []gin.HandlerFunc,
	userHandler *web.UserHandler, wechatHandler *web.OAuth2WechatHandler, articleHandler *web.ArticleHandler) *gin.Engine {
	r := gin.Default()
	r.Use(mdls...)
	//注册用户路由
	userHandler.RegisterRoutes(r)
	wechatHandler.RegisterRoutes(r)
	articleHandler.RegisterRoutes(r)
	return r
}

func InitMiddlewares(jwtHdl jwt2.JwtHandler, l logger.LoggerV1) []gin.HandlerFunc {
	//web请求日志打印 配置
	/*bd := middleware.NewLogMiddlewareBuilder(func(ctx context.Context, al *middleware.AccessLog) {
		l.Debug("HTTP请求", logger.Field{Key: "al", Val: al})
	}).AllowReqBody(true).AllowRespBody(true)
	viper.OnConfigChange(func(in fsnotify.Event) {
		ok := viper.GetBool("web.logreq")
		bd.AllowReqBody(ok)
	})*/
	//使用promethus统计
	pd := &metric.Builder{
		Namespace:  "lll",
		Subsystem:  "webook",
		Name:       "gin_http",
		Help:       "统计gin的http接口",
		InstanceId: "my-instance-1",
	}

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
				if strings.Contains(origin, "192.168.83.1") {
					return true
				}
				return strings.Contains(origin, "yourcompany.com")
			},
			//你的开发环境，包含localhost就允许访问
			MaxAge: 12 * time.Hour,
		}),

		//日志打印请求和响应
		//bd.Build(),

		//使用promethus统计
		pd.BuildResponseTime(),
		pd.BuildActiveRequest(),
		//使用OpenTelemetry自带的中间件
		otelgin.Middleware("webook"),

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

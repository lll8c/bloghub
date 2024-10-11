//go:build wireinject

package startup

import (
	"geektime/webook/internal/events/article"
	"geektime/webook/internal/repository"
	"geektime/webook/internal/repository/cache"
	"geektime/webook/internal/repository/dao"
	"geektime/webook/internal/service"
	"geektime/webook/internal/web"
	"geektime/webook/internal/web/jwt"
	"geektime/webook/ioc"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

// service
var thirdPartySet = wire.NewSet( // 第三方依赖
	InitRedis, InitDB,
	InitLogger)

var jobProviderSet = wire.NewSet(
	service.NewCronJobService,
	repository.NewPreemptJobRepository,
	dao.NewGORMJobDAO)

var userSvcProvider = wire.NewSet(
	dao.NewUserDao,
	cache.NewUserCache,
	repository.NewUserRepository,
	service.NewUserService)

var articlSvcProvider = wire.NewSet(
	repository.NewArticleRepository,
	cache.NewArticleRedisCache,
	dao.NewGROMArticleDAO,
	service.NewArticleService)

func InitArticleHandler(dao dao.ArticleDAO) *web.ArticleHandler {
	wire.Build(
		thirdPartySet,
		userSvcProvider,
		article.NewKafkaProducer,
		repository.NewArticleRepository,
		cache.NewArticleRedisCache,
		service.NewArticleService,
		web.NewArticleHandler)
	return &web.ArticleHandler{}
}

func InitWebServer() *gin.Engine {
	wire.Build(
		thirdPartySet,
		userSvcProvider,
		articlSvcProvider,
		// cache 部分
		cache.NewCodeCache,
		// repository 部分
		repository.NewCodeRepository,

		article.NewKafkaProducer,

		// Service 部分
		ioc.InitSMSService,
		service.NewCodeService,
		ioc.InitWechatService,

		// handler 部分
		web.NewUserHandler,
		web.NewArticleHandler,
		web.NewOAuth2WechatHandler,
		jwt.NewRedisJWTHandler,
		ioc.InitMiddlewares,
		ioc.InitWebServer,
	)
	return gin.Default()
}

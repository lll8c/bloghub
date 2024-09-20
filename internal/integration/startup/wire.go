//go:build wireinject

package startup

import (
	"geektime/webook/internal/repository"
	"geektime/webook/internal/repository/article"
	"geektime/webook/internal/repository/cache"
	"geektime/webook/internal/repository/dao"
	article2 "geektime/webook/internal/repository/dao/article"
	"geektime/webook/internal/service"
	"geektime/webook/internal/web"
	jwt2 "geektime/webook/internal/web/jwt"
	"geektime/webook/ioc"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

var thirdPartySet = wire.NewSet( // 第三方依赖
	InitRedis, InitDB,
	InitLogger)

var userSvcProvider = wire.NewSet(
	dao.NewUserDao,
	cache.NewUserCache,
	repository.NewUserRepository,
	service.NewUserService)

var articlSvcProvider = wire.NewSet(
	article.NewArticleRepository,
	service.NewArticleService)

func InitArticleHandler(articleDao article2.ArticleDAO) *web.ArticleHandler {
	wire.Build(
		thirdPartySet,
		article.NewArticleRepository,
		service.NewArticleService,
		web.NewArticleHandler)
	return &web.ArticleHandler{}
}

func InitWebServer() *gin.Engine {
	wire.Build(
		//第三方依赖
		ioc.InitDB, ioc.InitRedis,
		ioc.InitLoggerV1,
		//dao
		dao.NewUserDao, article2.NewGROMArticleDAO,
		//cache
		cache.NewUserCache, cache.NewCodeCache,
		//repository
		repository.NewUserRepository, repository.NewCodeRepository,
		article.NewArticleRepository,
		//service
		ioc.InitSMSService,
		ioc.InitWechatService,
		service.NewCodeService, service.NewUserService,
		service.NewArticleService,
		//handler
		jwt2.NewRedisJWTHandler,
		web.NewUserHandler,
		web.NewOAuth2WechatHandler,
		web.NewArticleHandler,
		ioc.InitMiddlewares,
		ioc.InitWebServer,
	)
	return new(gin.Engine)
}

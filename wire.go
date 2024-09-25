//go:build wireinject

package main

import (
	events "geektime/webook/internal/events/article"
	"geektime/webook/internal/repository"
	"geektime/webook/internal/repository/cache"
	"geektime/webook/internal/repository/dao"
	"geektime/webook/internal/service"
	"geektime/webook/internal/web"
	jwt2 "geektime/webook/internal/web/jwt"
	"geektime/webook/ioc"
	"github.com/google/wire"
)

func InitApp() *App {
	wire.Build(
		//第三方依赖
		ioc.InitDB, ioc.InitRedis,
		ioc.InitLoggerV1,

		//dao
		dao.NewUserDao,
		dao.NewGROMArticleDAO,
		dao.NewGORMInteractiveDAO,
		//cache
		cache.NewUserCache, cache.NewCodeCache,
		cache.NewArticleRedisCache,
		cache.NewInteractiveRedisCache,
		//repository
		repository.NewUserRepository, repository.NewCodeRepository,
		repository.NewArticleRepository,
		repository.NewCachedInteractiveRepository,
		//service
		ioc.InitSMSService,
		ioc.InitWechatService,
		service.NewCodeService, service.NewUserService,
		service.NewArticleService,
		service.NewInteractiveService,
		//handler
		jwt2.NewRedisJWTHandler,
		web.NewUserHandler,
		web.NewOAuth2WechatHandler,
		web.NewArticleHandler,
		ioc.InitMiddlewares,
		ioc.InitWebServer,
		//kafka, consumer and producer
		ioc.InitSaramaClient,
		ioc.InitSyncProducer,
		events.NewKafkaProducer,
		events.NewInteractiveReadEventConsumer,
		ioc.InitConsumers,
		//组装App结构体的所有字段
		wire.Struct(new(App), "*"),
	)
	return new(App)
}

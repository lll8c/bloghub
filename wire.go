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

var rankingSvcSet = wire.NewSet(
	cache.NewRankingRedisCache,
	repository.NewCachedRankingRepository,
	service.NewBatchRankingService,
)

func InitApp() *App {
	wire.Build(
		//第三方依赖
		ioc.InitDB, ioc.InitRedis,
		ioc.InitLoggerV1,
		ioc.InitRlockClient,
		//dao
		dao.NewUserDao,
		dao.NewGROMArticleDAO,
		//cache
		cache.NewUserCache, cache.NewCodeCache,
		cache.NewArticleRedisCache,
		//repository
		repository.NewUserRepository, repository.NewCodeRepository,
		repository.NewArticleRepository,
		//service
		ioc.InitSMSService,
		ioc.InitWechatService,
		service.NewCodeService, service.NewUserService,
		service.NewArticleService,
		//GRPC client
		ioc.InitEtcd,
		ioc.InitIntrGRPCClientV1,
		//handler
		jwt2.NewRedisJWTHandler,
		web.NewUserHandler,
		web.NewOAuth2WechatHandler,
		web.NewArticleHandler,
		ioc.InitMiddlewares,
		ioc.InitWebServer,
		//job
		rankingSvcSet,
		ioc.InitRankingJob,
		ioc.InitJobs,
		//kafka, consumer and producer
		ioc.InitKafkaClient,
		ioc.InitSyncProducer,
		events.NewKafkaProducer,

		//组装App结构体的所有字段
		wire.Struct(new(App), "*"),
	)
	return new(App)
}

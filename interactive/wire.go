//go:build wireinject

package main

import (
	"geektime/webook/interactive/events"
	"geektime/webook/interactive/grpc"
	"geektime/webook/interactive/ioc"
	"geektime/webook/interactive/repository"
	"geektime/webook/interactive/repository/cache"
	"geektime/webook/interactive/repository/dao"
	"geektime/webook/interactive/service"
	"github.com/google/wire"
)

var thirdPartySet = wire.NewSet( // 第三方依赖
	ioc.InitRedis,
	//双写数据库
	ioc.InitDstDB,
	ioc.InitSrcDB,
	ioc.InitDoubleWritePool,
	ioc.InitBizDB,

	ioc.InitKafkaClient,
	ioc.InitSyncProducer,
	ioc.InitLoggerV1,
)

var interactiveSvcSet = wire.NewSet(dao.NewGORMInteractiveDAO,
	cache.NewInteractiveRedisCache,
	repository.NewCachedInteractiveRepository,
	service.NewInteractiveService,
)

//不停机数据迁移后台管理服务端
//源目数据库进行校验，pool进行双写
var migratorProvider = wire.NewSet(
	ioc.InitInteractiveProducer,
	ioc.InitMigratorWebServer,
	ioc.InitFixerConsumer,
)

func InitApp() *App {
	wire.Build(
		thirdPartySet, interactiveSvcSet,
		//grpc服务端
		grpc.NewInteractiveServiceServer,
		ioc.InitGRPCxServer,
		events.NewInteractiveReadEventConsumer,
		migratorProvider,
		ioc.InitConsumers,
		//组装App结构体的所有字段
		wire.Struct(new(App), "*"),
	)
	return new(App)
}

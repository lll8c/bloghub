//go:build wireinject

package main

import (
	"geektime/webook/reward/grpc"
	"geektime/webook/reward/ioc"
	"geektime/webook/reward/repository"
	"geektime/webook/reward/repository/cache"
	"geektime/webook/reward/repository/dao"
	"geektime/webook/reward/service"
	"github.com/google/wire"
)

var thirdPartySet = wire.NewSet(
	ioc.InitDB,
	ioc.InitLogger,
	ioc.InitEtcdClient,
	ioc.InitRedis)

func Init() *App {
	wire.Build(thirdPartySet,
		service.NewWechatNativeRewardService,
		ioc.InitAccountClient,
		ioc.InitGRPCxServer,
		ioc.InitPaymentClient,
		repository.NewRewardRepository,
		cache.NewRewardRedisCache,
		dao.NewRewardGORMDAO,
		grpc.NewRewardServiceServer,
		wire.Struct(new(App), "GRPCServer"),
	)
	return new(App)
}

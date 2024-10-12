//go:build wireinject

package main

import (
	grpc2 "geektime/webook/follow/grpc"
	"geektime/webook/follow/ioc"
	"geektime/webook/follow/repository"
	"geektime/webook/follow/repository/dao"
	"geektime/webook/follow/service"
	"github.com/google/wire"
)

var serviceProviderSet = wire.NewSet(
	dao.NewGORMFollowRelationDAO,
	repository.NewFollowRelationRepository,
	service.NewFollowRelationService,
	grpc2.NewFollowRelationServiceServer,
)

var thirdProvider = wire.NewSet(
	ioc.InitDB,
	ioc.InitLogger,
)

func Init() *App {
	wire.Build(
		thirdProvider,
		serviceProviderSet,
		ioc.InitGRPCxServer,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}

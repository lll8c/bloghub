//go:build wireinject

package main

import (
	grpc2 "geektime/webook/comment/grpc"
	"geektime/webook/comment/ioc"
	"geektime/webook/comment/repository"
	"geektime/webook/comment/repository/dao"
	"geektime/webook/comment/service"
	"github.com/google/wire"
)

var serviceProviderSet = wire.NewSet(
	dao.NewCommentDAO,
	repository.NewCommentRepo,
	service.NewCommentSvc,
	grpc2.NewGrpcServer,
)

var thirdProvider = wire.NewSet(
	ioc.InitLogger,
	ioc.InitDB,
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

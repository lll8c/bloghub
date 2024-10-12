//go:build wireinject

package main

import (
	"geektime/webook/account/grpc"
	"geektime/webook/account/ioc"
	"geektime/webook/account/repository"
	"geektime/webook/account/repository/dao"
	"geektime/webook/account/service"
	"geektime/webook/pkg/wego"
	"github.com/google/wire"
)

func Init() *wego.App {
	wire.Build(
		ioc.InitDB,
		ioc.InitLogger,
		ioc.InitEtcdClient,
		ioc.InitGRPCxServer,
		dao.NewCreditGORMDAO,
		repository.NewAccountRepository,
		service.NewAccountService,
		grpc.NewAccountServiceServer,
		wire.Struct(new(wego.App), "GRPCServer"))
	return new(wego.App)
}

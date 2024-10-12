//go:build wireinject

package main

import (
	"geektime/webook/payment/grpc"
	"geektime/webook/payment/ioc"
	"geektime/webook/payment/repository"
	"geektime/webook/payment/repository/dao"
	"geektime/webook/payment/web"
	"github.com/google/wire"
)

func InitApp() *App {
	wire.Build(
		ioc.InitEtcdClient,
		ioc.InitKafka,
		ioc.InitProducer,
		ioc.InitWechatClient,
		dao.NewPaymentGORMDAO,
		ioc.InitDB,
		repository.NewPaymentRepository,
		grpc.NewWechatServiceServer,
		ioc.InitWechatNativeService,
		ioc.InitWechatConfig,
		ioc.InitWechatNotifyHandler,
		ioc.InitGRPCServer,
		web.NewWechatHandler,
		ioc.InitGinServer,
		ioc.InitLogger,
		wire.Struct(new(App), "WebServer", "GRPCServer"))
	return new(App)
}

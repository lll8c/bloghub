// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

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

// Injectors from wire.go:

func Init() *App {
	client := ioc.InitEtcdClient()
	wechatPaymentServiceClient := ioc.InitPaymentClient(client)
	db := ioc.InitDB()
	rewardDAO := dao.NewRewardGORMDAO(db)
	cmdable := ioc.InitRedis()
	rewardCache := cache.NewRewardRedisCache(cmdable)
	rewardRepository := repository.NewRewardRepository(rewardDAO, rewardCache)
	loggerV1 := ioc.InitLogger()
	accountServiceClient := ioc.InitAccountClient(client)
	rewardService := service.NewWechatNativeRewardService(wechatPaymentServiceClient, rewardRepository, loggerV1, accountServiceClient)
	rewardServiceServer := grpc.NewRewardServiceServer(rewardService)
	server := ioc.InitGRPCxServer(rewardServiceServer, client, loggerV1)
	app := &App{
		GRPCServer: server,
	}
	return app
}

// wire.go:

var thirdPartySet = wire.NewSet(ioc.InitDB, ioc.InitLogger, ioc.InitEtcdClient, ioc.InitRedis)

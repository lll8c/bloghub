//go:build wireinject

package main

import (
	"geektime/webook/internal/repository"
	"geektime/webook/internal/repository/cache"
	"geektime/webook/internal/repository/dao"
	"geektime/webook/internal/service"
	"geektime/webook/internal/web"
	"geektime/webook/ioc"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

func InitWebServer() *gin.Engine {
	wire.Build(
		//第三方依赖
		ioc.InitDB, ioc.InitRedis,
		//dao
		dao.NewUserDao,
		//cache
		cache.NewUserCache, cache.NewCodeCache,
		//repository
		repository.NewUserRepository, repository.NewCodeRepository,
		//service
		ioc.InitSMSService,
		service.NewCodeService, service.NewUserService,
		//handler
		web.NewUserHandler,
		ioc.InitMiddlewares,
		ioc.InitWebServer,
	)
	return new(gin.Engine)
}

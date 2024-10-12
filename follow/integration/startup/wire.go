//go:build wireinject

package startup

import (
	"geektime/webook/follow/grpc"
	"geektime/webook/follow/repository"
	"geektime/webook/follow/repository/cache"
	"geektime/webook/follow/repository/dao"
	"geektime/webook/follow/service"
	"github.com/google/wire"
)

func InitServer() *grpc.FollowServiceServer {
	wire.Build(
		InitRedis,
		InitLog,
		InitTestDB,
		dao.NewGORMFollowRelationDAO,
		cache.NewRedisFollowCache,
		repository.NewFollowRelationRepository,
		service.NewFollowRelationService,
		grpc.NewFollowRelationServiceServer,
	)
	return new(grpc.FollowServiceServer)
}

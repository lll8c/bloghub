//go:build wireinject

package startup

import (
	grpc2 "geektime/webook/comment/grpc"
	"geektime/webook/comment/repository"
	"geektime/webook/comment/repository/dao"
	"geektime/webook/comment/service"
	"geektime/webook/pkg/logger"
	"github.com/google/wire"
)

var serviceProviderSet = wire.NewSet(
	dao.NewCommentDAO,
	repository.NewCommentRepo,
	service.NewCommentSvc,
	grpc2.NewGrpcServer,
)

var thirdProvider = wire.NewSet(
	logger.NewNoOpLogger,
	InitTestDB,
)

func InitGRPCServer() *grpc2.CommentServiceServer {
	wire.Build(thirdProvider, serviceProviderSet)
	return new(grpc2.CommentServiceServer)
}

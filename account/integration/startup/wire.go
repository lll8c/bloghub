//go:build wireinject

package startup

import (
	"geektime/webook/account/grpc"
	"geektime/webook/account/repository"
	"geektime/webook/account/repository/dao"
	"geektime/webook/account/service"
	"github.com/google/wire"
)

func InitAccountService() *grpc.AccountServiceServer {
	wire.Build(InitTestDB,
		dao.NewCreditGORMDAO,
		repository.NewAccountRepository,
		service.NewAccountService,
		grpc.NewAccountServiceServer)
	return new(grpc.AccountServiceServer)
}

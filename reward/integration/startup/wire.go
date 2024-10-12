//go:build wireinject

package startup

import (
	pmtv1 "geektime/webook/api/proto/gen/payment/v1"
	"geektime/webook/reward/repository"
	"geektime/webook/reward/repository/cache"
	"geektime/webook/reward/repository/dao"
	"geektime/webook/reward/service"
	"github.com/google/wire"
)

var thirdPartySet = wire.NewSet(InitTestDB, InitLogger, InitRedis)

func InitWechatNativeSvc(client pmtv1.WechatPaymentServiceClient) *service.WechatNativeRewardService {
	wire.Build(service.NewWechatNativeRewardService,
		thirdPartySet,
		cache.NewRewardRedisCache,
		repository.NewRewardRepository, dao.NewRewardGORMDAO)
	return new(service.WechatNativeRewardService)
}

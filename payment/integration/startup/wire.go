//go:build wireinject

package startup

import (
	"geektime/webook/payment/ioc"
	"geektime/webook/payment/repository"
	"geektime/webook/payment/repository/dao"
	"geektime/webook/payment/service/wechat"
	"github.com/google/wire"
)

var thirdPartySet = wire.NewSet(ioc.InitLogger, InitTestDB)

var wechatNativeSvcSet = wire.NewSet(
	ioc.InitWechatClient,
	dao.NewPaymentGORMDAO,
	repository.NewPaymentRepository,
	ioc.InitWechatNativeService,
	ioc.InitWechatConfig)

func InitWechatNativeService() *wechat.NativePaymentService {
	wire.Build(wechatNativeSvcSet, thirdPartySet)
	return new(wechat.NativePaymentService)
}

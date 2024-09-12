package ratelimit

import (
	"context"
	"fmt"
	"geektime/webook/internal/service/sms"
	"geektime/webook/pkg/ratelimit"
)

var (
	errLimited = fmt.Errorf("发生了限流")
)

// LimitSMSService 装饰器模式使用发送短信服务
// 采用了redis滑动窗口限流
type LimitSMSService struct {
	svc     sms.Service
	limiter ratelimit.Limiter
}

func NewLimitSMSService(svc sms.Service, limiter ratelimit.Limiter) *LimitSMSService {
	return &LimitSMSService{
		svc:     svc,
		limiter: limiter,
	}
}

// Send 装饰这个方法
func (s LimitSMSService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	//加一些代码，新特性
	//使用redis限流功能
	limited, err := s.limiter.Limit(ctx, "sms:tencent")
	//系统错误
	if err != nil {
		//可以限流: 保守策略；下游很坑的时候
		//可以不限: 下游很强，或业务可用性要求很高，尽量容错策略

		//这里采用保守策略
		//包装错误
		return fmt.Errorf("短信服务判断是否限流出现错误 %w", err)
	}
	//发生了限流
	if limited {
		return errLimited
	}
	err = s.svc.Send(ctx, tplId, args, numbers...)
	return err
}

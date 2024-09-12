package retryable

import (
	"context"
	"errors"
	"geektime/webook/internal/service/sms"
)

// Service 提供自动重试功能
type Service struct {
	svc sms.Service
	//重试次数
	retryMax int
}

func (s Service) Send(ctx context.Context, tpl string, args []string, numbers ...string) error {
	err := s.svc.Send(ctx, tpl, args, numbers...)
	//当前重试次数
	cnt := 1
	if err != nil && cnt < s.retryMax {
		err = s.svc.Send(ctx, tpl, args, numbers...)
		if err == nil {
			return nil
		}
		cnt++
	}
	return errors.New("所有重试都失败了")
}

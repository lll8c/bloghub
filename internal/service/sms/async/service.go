package async

import (
	"context"
	"geektime/webook/internal/service/sms"
)

type SMSService struct {
	svc sms.Service
}

func NewSMSService() *SMSService {
	return &SMSService{}
}

func (s SMSService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	//首先是正常路径
	err := s.svc.Send(ctx, tplId, args, numbers...)
	if err != nil {
		//判定是不是崩溃了

	}

	return nil
}

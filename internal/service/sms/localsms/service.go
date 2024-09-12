package localsms

import (
	"context"
	"log"
)

// Service 基于本地实现，模拟发短信服务
type Service struct {
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	log.Println("验证码是", args)
	return nil
}

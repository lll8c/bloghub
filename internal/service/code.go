package service

import (
	"context"
	"fmt"
	"geektime/webook/internal/repository"
	"geektime/webook/internal/service/sms"
	"math/rand"
)

var (
	ErrCodeVerifyTooMany = repository.ErrCodeVerifyTooMany
	ErrCodeSendTooMany   = repository.ErrCodeSendTooMany
)

type CodeService interface {
	// Send bix 区别业务场景，用于什么的验证码
	Send(ctx context.Context, biz, phone string) error
	Verify(ctx context.Context,
		biz, phone, inputCode string) (bool, error)
}

type codeService struct {
	repo repository.CodeRepository
	sms  sms.Service
}

func NewCodeService(repo repository.CodeRepository, smsSvc sms.Service) CodeService {
	return &codeService{
		repo: repo,
		sms:  smsSvc,
	}
}

func (svc *codeService) Send(ctx context.Context, biz, phone string) error {
	//生成验证码
	code := svc.generate()
	//存入redis
	err := svc.repo.Set(ctx, biz, phone, code)
	if err != nil {
		return err
	}
	//通过手机短信服务发送验证码
	return svc.sms.Send(ctx, []string{code}, phone)
}

func (svc *codeService) Verify(ctx context.Context,
	biz, phone, inputCode string) (bool, error) {
	ok, err := svc.repo.Verify(ctx, biz, phone, inputCode)
	if err == repository.ErrCodeVerifyTooMany {
		// 相当于，我们对外面屏蔽了验证次数过多的错误，我们就是告诉调用者，你这个不对
		return false, nil
	}
	return ok, err
}

func (svc *codeService) generate() string {
	// 0-999999
	code := rand.Intn(1000000)
	return fmt.Sprintf("%06d", code)
}

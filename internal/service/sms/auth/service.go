package auth

import (
	"context"
	"errors"
	"geektime/webook/internal/service/sms"
	"github.com/golang-jwt/jwt/v5"
)

// SMSService 使用功能token提高安全性
type SMSService struct {
	svc sms.Service
	key []byte
}

func (s *SMSService) Send(ctx context.Context, tplToken string, args []string, numbers ...string) error {
	var claims SMSClaims
	//如果这里能解析成功 且 token有效，就说明是对应的业务方
	token, err := jwt.ParseWithClaims(tplToken, &claims, func(token *jwt.Token) (interface{}, error) {
		return s.key, nil
	})
	if err != nil {
		return err
	}
	if !token.Valid {
		return errors.New("token 不合法")
	}
	return s.svc.Send(ctx, claims.Tpl, args, numbers...)
}

type SMSClaims struct {
	jwt.RegisteredClaims
	Tpl string
	// 额外加字段
}

package service

import (
	"context"
	"errors"
	"geektime/webook/internal/domain"
	"geektime/webook/internal/repository"
	svcmocks "geektime/webook/internal/repository/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
	"testing"
)

func Test_userService_Login(t *testing.T) {
	var testCase = []struct {
		name     string
		mock     func(ctrl *gomock.Controller) repository.UserRepository
		ctx      context.Context
		email    string
		password string
		wantUser domain.User
		wantErr  error
	}{
		{
			name: "登录成功",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := svcmocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").
					Return(domain.User{Email: "123@qq.com", Password: "$2a$10$HGuKZIFj3zZS.ZDcG1ABq.IbbQhnEV2Kbybf8jnHkizCnDn15c0bO"}, nil)
				return repo
			},
			email:    "123@qq.com",
			password: "hello@123",
			wantUser: domain.User{
				Email:    "123@qq.com",
				Password: "$2a$10$HGuKZIFj3zZS.ZDcG1ABq.IbbQhnEV2Kbybf8jnHkizCnDn15c0bO",
			},
			wantErr: nil,
		},
		{
			name: "用户不存在",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := svcmocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").
					Return(domain.User{}, repository.ErrUserNotFound)
				return repo
			},
			email:    "123@qq.com",
			password: "hello@123",
			wantUser: domain.User{},
			wantErr:  ErrInvalidUserOrPassword,
		},
		{
			name: "DB错误",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := svcmocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").
					Return(domain.User{}, errors.New("mock db 错误"))
				return repo
			},
			email:    "123@qq.com",
			password: "hello@123",
			wantUser: domain.User{},
			wantErr:  errors.New("mock db 错误"),
		},
		{
			name: "密码不对",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := svcmocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").
					Return(domain.User{Email: "123@qq.com", Password: "$2a$10$HGuewKZIFj3zZS.ZDcG1ABq.IbbQhnEV2Kbf8jnHkizCnDn15c0bO"}, nil)
				return repo
			},
			email:    "123@qq.com",
			password: "hello@123",
			wantUser: domain.User{},
			wantErr:  ErrInvalidUserOrPassword,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := NewUserService(tc.mock(ctrl))
			u, err := svc.Login(tc.ctx, tc.email, tc.password)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, u)
		})
	}
}

// 生成加密后的密码
func TestEncrypted(t *testing.T) {
	res, err := bcrypt.GenerateFromPassword([]byte("hello@123"), bcrypt.DefaultCost)
	if err == nil {
		t.Log(string(res))
	}
}

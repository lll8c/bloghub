package web

import (
	"bytes"
	"context"
	"errors"
	"geektime/webook/internal/domain"
	"geektime/webook/internal/service"
	svcmocks "geektime/webook/internal/service/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test(t *testing.T) {
	zap.L().Debug("msg", zap.String("name", "zs"))
}

func TestUserHandler_SignUp(t *testing.T) {
	//测试用例
	var testCase = []struct {
		name    string
		mock    func(ctrl *gomock.Controller) service.UserService
		reqBody string

		wantCode int
		wantBody string
	}{
		{
			name: "注册成功",
			mock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmocks.NewMockUserService(ctrl)
				usersvc.EXPECT().SignUp(gomock.Any(), domain.User{
					Email:    "123@qq.com",
					Password: "hello@123",
				}).Return(nil)
				return usersvc
			},
			reqBody: `{
				"email": "123@qq.com",
				"password": "hello@123",
				"confirm_password": "hello@123"
				}`,
			wantCode: 200,
			wantBody: "注册成功",
		},
		{
			name: "绑定参数不对",
			mock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmocks.NewMockUserService(ctrl)
				return usersvc
			},
			reqBody: `
				"email1ello@123nfirm_passwllo@123"
				}`,
			wantCode: http.StatusBadRequest,
			wantBody: "",
		},
		{
			name: "邮箱格式不对",
			mock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmocks.NewMockUserService(ctrl)
				return usersvc
			},
			reqBody: `{
				"email": "123",
				"password": "hello@123",
				"confirm_password": "hello@123"
				}`,
			wantCode: 200,
			wantBody: "你的邮箱格式不对",
		},
		{
			name: "两次密码不匹配",
			mock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmocks.NewMockUserService(ctrl)
				return usersvc
			},
			reqBody: `{
				"email": "123@qq.com",
				"password": "hello@1231324",
				"confirm_password": "hello@123"
				}`,
			wantCode: 200,
			wantBody: "两次输入的密码不一致",
		},
		{
			name: "密码格式不对",
			mock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmocks.NewMockUserService(ctrl)
				return usersvc
			},
			reqBody: `{
				"email": "123@qq.com",
				"password": "hello123",
				"confirm_password": "hello123"
				}`,
			wantCode: 200,
			wantBody: "密码必须大于8位，包含数字、字母、特殊字符",
		},
		{
			name: "系统异常",
			mock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmocks.NewMockUserService(ctrl)
				usersvc.EXPECT().SignUp(gomock.Any(), domain.User{
					Email:    "123@qq.com",
					Password: "hello@123",
				}).Return(errors.New("随便一个error"))
				return usersvc
			},
			reqBody: `{
				"email": "123@qq.com",
				"password": "hello@123",
				"confirm_password": "hello@123"
				}`,
			wantCode: 200,
			wantBody: "系统异常",
		},
		{
			name: "邮箱冲突",
			mock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmocks.NewMockUserService(ctrl)
				usersvc.EXPECT().SignUp(gomock.Any(), domain.User{
					Email:    "123@qq.com",
					Password: "hello@123",
				}).Return(service.ErrUserDuplicate)
				return usersvc
			},
			reqBody: `{
				"email": "123@qq.com",
				"password": "hello@123",
				"confirm_password": "hello@123"
				}`,
			wantCode: 200,
			wantBody: "邮箱冲突",
		},
	}
	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			//先创建一个mock控制器
			ctrl := gomock.NewController(t)
			//每个测试结束都要调用Finish
			//然后mock就会验证测试流程是否符合预期
			defer ctrl.Finish()

			server := gin.Default()
			//创建模拟对象，并设置模拟调用
			h := NewUserHandler(tc.mock(ctrl), nil)
			h.RegisterRoutes(server)
			//构造http请求
			req, err := http.NewRequest(http.MethodPost, "/users/signup",
				bytes.NewBuffer([]byte(tc.reqBody)))
			require.NoError(t, err)
			//数据是json格式
			req.Header.Set("Content-Type", "application/json")

			//初始化一个响应体
			resp := httptest.NewRecorder()
			//gin会处理这个请求
			//响应写回到resp里
			server.ServeHTTP(resp, req)

			t.Log(resp)
			assert.Equal(t, tc.wantCode, resp.Code)
			assert.Equal(t, tc.wantBody, resp.Body.String())
		})
	}
}

func TestMock(t *testing.T) {
	//先创建一个mock的控制器
	ctrl := gomock.NewController(t)
	//每个测试结束都要调用Finish
	//然后mock就会验证你的验证流程是否符合预期
	defer ctrl.Finish()
	usersvc := svcmocks.NewMockUserService(ctrl)
	//开始设计一个模拟调用
	//预期是一个Signup的调用
	//模拟的条件是gomock.Any,gomock.Any
	//然后返回
	usersvc.EXPECT().SignUp(gomock.Any(), domain.User{
		Email: "123@qq.com",
	}).Return(errors.New("mock error"))

	err := usersvc.SignUp(context.Background(), domain.User{
		Email: "123@qq.com",
	})
	t.Log(err)
}

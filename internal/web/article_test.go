package web

import (
	"bytes"
	"encoding/json"
	"errors"
	"geektime/webook/internal/domain"
	"geektime/webook/internal/service"
	svcmocks "geektime/webook/internal/service/mocks"
	jwt2 "geektime/webook/internal/web/jwt"
	"geektime/webook/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestArticleHandler_Publish(t *testing.T) {
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) service.ArticleService

		reqBody  string
		wantCode int
		wantRes  Result
	}{
		{
			name: "新建并且发表成功",
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				svc := svcmocks.NewMockArticleService(ctrl)
				svc.EXPECT().Publish(gomock.Any(), domain.Article{
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(1), nil)
				return svc
			},
			reqBody: `
				{
				 "title": "我的标题",
				 "content": "我的内容"
				}
				`,
			wantCode: 200,
			wantRes: Result{
				// 原本是 int64的，但是因为 Data 是any，所以在反序列化的时候，
				// 用的 float64
				Msg:  "Ok",
				Data: float64(1),
			},
		},
		{
			"publish失败",
			func(ctrl *gomock.Controller) service.ArticleService {
				svc := svcmocks.NewMockArticleService(ctrl)
				svc.EXPECT().Publish(gomock.Any(), domain.Article{
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(0), errors.New("publish识别"))
				return svc
			},
			`
				{
				 "title": "我的标题",
				 "content": "我的内容"
				}
				`,
			200,
			Result{
				// 原本是 int64的，但是因为 Data 是any，所以在反序列化的时候，
				// 用的 float64
				Msg:  "Ok",
				Data: float64(1),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// 构造 handler
			svc := tc.mock(ctrl)
			hdl := NewArticleHandler(svc, &logger.NopLogger{})

			// 准备服务器，注册路由
			server := gin.Default()
			server.Use(func(ctx *gin.Context) {
				ctx.Set("claims", &jwt2.UserClaims{
					Uid: 123,
				})
			})
			hdl.RegisterRoutes(server)

			// 准备Req和记录的 recorder
			req, err := http.NewRequest(http.MethodPost,
				"/articles/publish", bytes.NewReader([]byte(tc.reqBody)))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()

			// 执行
			server.ServeHTTP(recorder, req)
			assert.Equal(t, tc.wantCode, recorder.Code)
			if recorder.Code != http.StatusOK {
				return
			}
			var res Result
			err = json.NewDecoder(recorder.Body).Decode(&res)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantRes, res)
		})
	}
}

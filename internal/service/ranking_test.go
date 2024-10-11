//go:build need_fix

package service

import (
	"context"
	intrv1 "geektime/webook/api/proto/gen/intr/v1"
	domain2 "geektime/webook/interactive/domain"
	"geektime/webook/interactive/service"
	"geektime/webook/internal/domain"
	svcmocks "geektime/webook/internal/service/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

func TestRankingTopN(t *testing.T) {
	const batchSize = 2
	now := time.Now()
	testCases := []struct {
		name string

		mock func(ctrl *gomock.Controller) (service.InteractiveService, ArticleService)

		wantArts []domain.Article
		wantErr  error
	}{
		{
			name: "成功获取",
			mock: func(ctrl *gomock.Controller) (intrv1.InteractiveServiceClient, ArticleService) {
				intrSvc := svcmocks.NewMockInteractiveService(ctrl)
				artSvc := svcmocks.NewMockArticleService(ctrl)
				// 先模拟批量获取数据
				// 先模拟第一批
				artSvc.EXPECT().ListPub(gomock.Any(), 0, 3).
					Return([]domain.Article{
						{Id: 1, Utime: now},
						{Id: 2, Utime: now},
						{Id: 3, Utime: now},
					}, nil)
				// 模拟第二批
				artSvc.EXPECT().ListPub(gomock.Any(), 3, 3).
					Return([]domain.Article{
						{Id: 4, Utime: now},
						{Id: 5, Utime: now},
					}, nil)

				// 第一批的点赞数据
				intrSvc.EXPECT().GetByIds(gomock.Any(), "article", []int64{1, 2, 3}).
					Return(map[int64]domain2.Interactive{
						1: {LikeCnt: 1},
						2: {LikeCnt: 1},
						3: {LikeCnt: 2},
					}, nil)

				// 第二批的点赞数据
				intrSvc.EXPECT().GetByIds(gomock.Any(), "article", []int64{4, 5}).
					Return(map[int64]domain2.Interactive{
						4: {LikeCnt: 1},
						5: {LikeCnt: 4},
					}, nil)

				return intrSvc, artSvc
			},

			wantErr: nil,
			wantArts: []domain.Article{
				{Id: 5, Utime: now},
				{Id: 3, Utime: now},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			intrSvc, artSvc := tc.mock(ctrl)
			svc := NewBatchRankingService(artSvc, intrSvc)
			svc.batchSize = 3
			svc.n = 2
			arts, err := svc.topN(context.Background())
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantArts, arts)
		})
	}
}

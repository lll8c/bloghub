package service

import (
	"context"
	intrv1 "geektime/webook/api/proto/gen/intr/v1"
	"geektime/webook/internal/domain"
	"geektime/webook/internal/repository"
	"github.com/ecodeclub/ekit/queue"
	"github.com/ecodeclub/ekit/slice"
	"math"
	"time"
)

type RankingService interface {
	TopN(ctx context.Context) error
	GetTopN(ctx context.Context) ([]domain.Article, error)
}

type BatchRankingService struct {
	repo      repository.RankingRepository
	artSvc    ArticleService
	intrSvc   intrv1.InteractiveServiceClient
	batchSize int
	n         int
	scoreFunc func(t time.Time, likeCnt int64) float64
}

func NewBatchRankingService(repo repository.RankingRepository, artSvc ArticleService, intrSvc intrv1.InteractiveServiceClient) RankingService {
	return &BatchRankingService{
		repo:      repo,
		artSvc:    artSvc,
		intrSvc:   intrSvc,
		batchSize: 100,
		n:         100,
		scoreFunc: func(t time.Time, likeCnt int64) float64 {
			sec := time.Since(t).Seconds()
			return float64(likeCnt-1) / math.Pow(float64(sec+2), 1.5)
		},
	}
}

// GetTopN 查询热榜数据
func (b *BatchRankingService) GetTopN(ctx context.Context) ([]domain.Article, error) {
	return b.repo.GetTopN(ctx)
}

// TopN 更新热榜数据
func (b *BatchRankingService) TopN(ctx context.Context) error {
	//获取热度前100的数据
	arts, err := b.topN(ctx)
	if err != nil {
		return err
	}
	// 最终是要放到缓存里面的
	// 存到缓存里面
	return b.repo.ReplaceTopN(ctx, arts)
}

// 更新并获取热榜article
func (b *BatchRankingService) topN(ctx context.Context) ([]domain.Article, error) {
	//只取7天内的数据
	now := time.Now()
	//先拿一批数据
	offset := 0
	type Score struct {
		art   domain.Article
		score float64
	}
	topN := queue.NewConcurrentPriorityQueue[Score](b.n,
		func(src Score, dst Score) int {
			if src.score > dst.score {
				return 1
			} else if src.score == dst.score {
				return 0
			} else {
				return -1
			}
		})
	//先拿一批数据
	for {
		arts, err := b.artSvc.ListPub(ctx, now, offset, b.batchSize)
		if err != nil {
			return nil, err
		}
		//获取ids
		ids := slice.Map[domain.Article, int64](arts,
			func(idx int, src domain.Article) int64 {
				return src.Id
			})
		//找对应的点赞数据
		intrs, err := b.intrSvc.GetByIds(ctx, &intrv1.GetByIdsRequest{
			Biz: "article",
			Ids: ids,
		})
		//计算score
		for _, art := range arts {
			intr, ok := intrs.Intrs[art.Id]
			//没有对应的点赞数据
			if !ok {
				continue
			}
			score := b.scoreFunc(art.Utime, intr.LikeCnt)
			//队列未满直接入队
			err = topN.Enqueue(Score{
				art:   art,
				score: score,
			})
			//队列已满，考虑score在不在前100名，在就替换score最小的数据
			if err == queue.ErrOutOfCapacity {
				val, _ := topN.Dequeue()
				if val.score < score {
					topN.Enqueue(Score{
						art:   art,
						score: score,
					})
				}
			}
		}
		//如果没有下一批数据了，或已经到了7天前的数据
		if len(arts) < b.batchSize || now.Sub(arts[len(arts)-1].Utime).Hours() > 7*24 {
			break
		}
		offset = offset + len(arts)
	}
	//最后得出结果
	res := make([]domain.Article, b.n)
	for i := b.n - 1; i >= 0; i-- {
		val, err := topN.Dequeue()
		//取完了不够n个数据
		if err != nil {
			break
		}
		res[i] = val.art
	}
	return res, nil
}

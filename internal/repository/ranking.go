package repository

import (
	"context"
	"geektime/webook/internal/domain"
	"geektime/webook/internal/repository/cache"
)

type RankingRepository interface {
	ReplaceTopN(ctx context.Context, arts []domain.Article) error
	GetTopN(ctx context.Context) ([]domain.Article, error)
}

type CachedRankingRepository struct {
	cache cache.RankingCache
	// 下面是给 v1 用的
	redisCache *cache.RankingRedisCache
	localCache *cache.RankingLocalCache
}

// NewCachedRankingRepository 只用redis缓存
func NewCachedRankingRepository(cache cache.RankingCache) RankingRepository {
	return &CachedRankingRepository{cache: cache}
}

func (repo *CachedRankingRepository) GetTopN(ctx context.Context) ([]domain.Article, error) {
	return repo.cache.Get(ctx)
}

func (repo *CachedRankingRepository) ReplaceTopN(ctx context.Context, arts []domain.Article) error {
	return repo.cache.Set(ctx, arts)
}

// NewCachedRankingRepositoryV1 使用本地缓存已经redis缓存
func NewCachedRankingRepositoryV1(redisCache *cache.RankingRedisCache, localCache *cache.RankingLocalCache) *CachedRankingRepository {
	return &CachedRankingRepository{redisCache: redisCache, localCache: localCache}

}

func (repo *CachedRankingRepository) GetTopNV1(ctx context.Context) ([]domain.Article, error) {
	res, err := repo.localCache.Get(ctx)
	if err == nil {
		return res, nil
	}
	res, err = repo.redisCache.Get(ctx)
	//回写本地缓存
	if err == nil {
		repo.localCache.Set(ctx, res)
	} else {
		//使用本地缓存解决redis可用性问题，进行兜底
		return repo.localCache.ForceGet(ctx)
	}
	return res, err
}

func (repo *CachedRankingRepository) ReplaceTopNV1(ctx context.Context, arts []domain.Article) error {
	_ = repo.localCache.Set(ctx, arts)
	return repo.redisCache.Set(ctx, arts)
}

package repository

import (
	"context"
	"geektime/webook/internal/domain"
)

type ArticleReaderRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	Update(ctx context.Context, art domain.Article) error
	FindById(ctx context.Context, id int64) (domain.Article, error)
	// Save 有就更新，没有
	Save(ctx context.Context, art domain.Article) (int64, error)
}

type articleReaderRepository struct {
}

func (a *articleReaderRepository) Save(ctx context.Context, art domain.Article) (int64, error) {
	return 0, nil
}

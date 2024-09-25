package repository

import (
	"context"
	"geektime/webook/internal/domain"
	"geektime/webook/internal/repository/dao"
)

type ArticleAuthorRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	Update(ctx context.Context, art domain.Article) error
	FindById(ctx context.Context, id int64) (domain.Article, error)
}

type CacheArticleRepository struct {
	dao dao.ArticleDAO
}

func NewArticleAuthorRepository(dao dao.ArticleDAO) ArticleAuthorRepository {
	return &CacheArticleRepository{
		dao: dao,
	}
}

func (c *CacheArticleRepository) Create(ctx context.Context, art domain.Article) (int64, error) {
	return c.dao.Insert(ctx, dao.Article{
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
	})
}

func (c *CacheArticleRepository) Update(ctx context.Context, art domain.Article) error {
	return c.dao.Update(ctx, dao.Article{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
	})
}

func (c *CacheArticleRepository) FindById(ctx context.Context, id int64) (domain.Article, error) {
	art, err := c.dao.FindById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	return domain.Article{
		Id:      art.Id,
		Title:   art.Title,
		Content: art.Content,
		Author: domain.Author{
			Id: art.AuthorId,
		},
	}, nil
}

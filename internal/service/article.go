package service

import (
	"context"
	"geektime/webook/internal/domain"
	"geektime/webook/internal/repository/article"
	"geektime/webook/pkg/logger"
)

type ArticleService interface {
	Save(ctx context.Context, art domain.Article) (int64, error)
	Publish(ctx context.Context, art domain.Article) (int64, error)
	PublishV1(ctx context.Context, art domain.Article) (int64, error)
	Withdraw(ctx context.Context, art domain.Article) error
	GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetPubById(ctx context.Context, id int64) (domain.Article, error)
}

type articleService struct {
	repo article.ArticleRepository

	author article.ArticleAuthorRepository
	reader article.ArticleReaderRepository
	l      logger.LoggerV1
}

func NewArticleService(repo article.ArticleRepository, l logger.LoggerV1) ArticleService {
	return &articleService{
		repo: repo,
		l:    l,
	}
}

// Save 修改或者创建帖子，保存
func (a *articleService) Save(ctx context.Context, art domain.Article) (int64, error) {
	//将帖子的状态设置为未发表
	art.Status = domain.ArticleStatusUnpublished
	if art.Id > 0 {
		err := a.repo.Update(ctx, art)
		return art.Id, err
	}
	return a.repo.Create(ctx, art)
}

func (a *articleService) Publish(ctx context.Context, art domain.Article) (int64, error) {
	art.Status = domain.ArticleStatusPublished
	return a.repo.Sync(ctx, art)
}

// PublishV1 依靠连个不同repository来完成
func (a *articleService) PublishV1(ctx context.Context, art domain.Article) (int64, error) {
	var (
		id  = art.Id
		err error
	)
	if art.Id > 0 {
		err = a.author.Update(ctx, art)
	} else {
		id, err = a.author.Create(ctx, art)
	}
	if err != nil {
		return 0, err
	}
	//因为service层没有事务，所有为了保证两个操作同时成功，只能重试
	//制作库和线上库的ID是相等的
	//重试保存
	art.Id = id
	for i := 0; i < 3; i++ {
		id, err = a.reader.Save(ctx, art)
		if err == nil {
			break
		}
		a.l.Error("保存失败，保存到线上失败",
			logger.Int64("art_id", art.Id),
			logger.Error(err))
	}
	if err != nil {
		a.l.Error("部分失败，重试彻底失败",
			logger.Int64("art_id", art.Id),
			logger.Error(err))
		//接入告警系统，手工处理一下
		//走异步，直接保存到文件
		//走Canal
		//打MQ
	}
	return id, err
}

func (a *articleService) Withdraw(ctx context.Context, art domain.Article) error {
	return a.repo.SyncStatus(ctx, art.Id, art.Author.Id, domain.ArticleStatusPrivate)
}

func (a *articleService) GetPubById(ctx context.Context, id int64) (domain.Article, error) {
	return a.repo.GetPubById(ctx, id)
}

func (a *articleService) GetById(ctx context.Context, id int64) (domain.Article, error) {
	return a.repo.GetById(ctx, id)
}

func (a *articleService) GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	return a.repo.GetByAuthor(ctx, uid, offset, limit)
}

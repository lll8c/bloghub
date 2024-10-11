package repository

import (
	"context"
	"geektime/webook/internal/domain"
	"geektime/webook/internal/repository/cache"
	"geektime/webook/internal/repository/dao"
	"geektime/webook/pkg/logger"
	"github.com/ecodeclub/ekit/slice"
	"time"
)

type ArticleRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	Update(ctx context.Context, art domain.Article) error
	Sync(ctx context.Context, art domain.Article) (int64, error)
	// SyncV1 存储并同步数据
	SyncV1(ctx context.Context, art domain.Article) (int64, error)
	SyncStatus(ctx context.Context, artId int64, authorId int64, status int) error

	GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]domain.Article, error)
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetPubById(ctx context.Context, id int64) (domain.Article, error)
	preCache(ctx context.Context, arts []domain.Article)
}

type articleRepository struct {
	dao       dao.ArticleDAO
	readerDao dao.ReaderDao
	authorDao dao.AuthorDao
	userDao   dao.UserDAO

	cache cache.ArticleCache
	l     logger.LoggerV1
}

func NewArticleRepository(dao dao.ArticleDAO, cache cache.ArticleCache, l logger.LoggerV1, userDao dao.UserDAO) ArticleRepository {
	return &articleRepository{
		dao:     dao,
		cache:   cache,
		l:       l,
		userDao: userDao,
	}
}
func (c *articleRepository) ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]domain.Article, error) {
	arts, err := c.dao.ListPub(ctx, start, offset, limit)
	if err != nil {
		return nil, err
	}
	return slice.Map[dao.PublishArticle, domain.Article](arts,
		func(idx int, src dao.PublishArticle) domain.Article {
			return c.toDomain(dao.Article(src))
		}), nil
}

func (c *articleRepository) Sync(ctx context.Context, art domain.Article) (int64, error) {
	//发表可能会更新制作库，清空用户文章第一页缓存
	err := c.cache.DelFirstPage(ctx, art.Author.Id)
	if err != nil {
		c.l.Error("删除缓存失败", logger.Int64("authorId", art.Author.Id))
		return 0, err
	}
	//缓存发表文章
	err = c.cache.SetPub(ctx, art)
	if err != nil {
		c.l.Error("缓存线上库数据失败", logger.Int64("authorId", art.Author.Id))
		return 0, err
	}

	return c.dao.Sync(ctx, c.ToEntity(art))
}

func (c *articleRepository) SyncV1(ctx context.Context, art domain.Article) (int64, error) {
	var (
		id  = art.Id
		err error
	)
	//操作制作库
	if art.Id > 0 {
		err = c.authorDao.Update(ctx, c.ToEntity(art))
	} else {
		id, err = c.authorDao.Insert(ctx, c.ToEntity(art))
	}
	if err != nil {
		return id, err
	}
	//线上库可能有数据页可能没有数据
	//如果数据库有则更新，没有则插入
	err = c.readerDao.Upsert(ctx, c.ToEntity(art))
	return id, err
}

func (c *articleRepository) SyncStatus(ctx context.Context, artId int64, authorId int64, status int) error {
	//清空缓存
	err := c.cache.DelFirstPage(ctx, authorId)
	if err != nil {
		c.l.Error("删除缓存失败", logger.Int64("authorId", authorId))
		return err
	}
	return c.dao.SyncStatus(ctx, artId, authorId, status)
}

func (c *articleRepository) Create(ctx context.Context, art domain.Article) (int64, error) {
	//清空用户文章第一页缓存
	err := c.cache.DelFirstPage(ctx, art.Author.Id)
	if err != nil {
		c.l.Error("删除缓存失败", logger.Int64("authorId", art.Author.Id))
		return 0, err
	}
	return c.dao.Insert(ctx, c.ToEntity(art))
}

func (c *articleRepository) Update(ctx context.Context, art domain.Article) error {
	//清空缓存
	err := c.cache.DelFirstPage(ctx, art.Author.Id)
	if err != nil {
		c.l.Error("删除缓存失败", logger.Int64("authorId", art.Author.Id))
		return err
	}
	return c.dao.Update(ctx, c.ToEntity(art))
}

func (c *articleRepository) GetPubById(ctx context.Context, id int64) (domain.Article, error) {
	//读取缓存
	res, err := c.cache.GetPub(ctx, id)
	if err == nil {
		return res, err
	}
	//读取线上库数据，如果Content放到了OSS中，就要让前端去读取Content
	art, err := c.dao.GetPubById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	// 我现在要去查询对应的 User 信息，拿到创作者信息
	res = c.toDomain(dao.Article(art))
	author, err := c.userDao.FindById(ctx, art.AuthorId)
	if err != nil {
		c.l.Error("获取用户信息失败", logger.Error(err))
		return domain.Article{}, err
	}
	res.Author.Name = author.Nickname
	//回写缓存
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		er := c.cache.SetPub(ctx, res)
		if er != nil {
			// 记录日志
		}
	}()
	return res, nil
}

// GetById 查询用户帖子列表后，点击查询帖子详情
func (c *articleRepository) GetById(ctx context.Context, id int64) (domain.Article, error) {
	//查询缓存
	res, err := c.cache.Get(ctx, id)
	if err == nil {
		return res, nil
	}
	//查询数据库
	art, err := c.dao.GetById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	res = c.toDomain(art)
	//异步回写缓存
	go func() {
		er := c.cache.Set(ctx, res)
		if er != nil {
			// 记录日志
		}
	}()
	return res, nil
}

func (c *articleRepository) GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	// 首先第一步，判定要不要查询缓存
	// 事实上， limit <= 100 都可以查询缓存
	if offset == 0 && limit == 100 {
		//if offset == 0 && limit <= 100 {
		res, err := c.cache.GetFirstPage(ctx, uid)
		if err == nil {
			return res, err
		} else {
			// 要考虑记录日志
			// 缓存未命中，你是可以忽略的
		}
	}
	//查询数据库
	arts, err := c.dao.GetByAuthor(ctx, uid, offset, limit)
	if err != nil {
		return nil, err
	}
	res := slice.Map[dao.Article, domain.Article](arts, func(idx int, src dao.Article) domain.Article {
		return c.toDomain(src)
	})

	//回写缓存
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if offset == 0 && limit == 100 {
			// 缓存回写失败，不一定是大问题，但有可能是大问题
			err = c.cache.SetFirstPage(ctx, uid, res)
			if err != nil {
				// 记录日志
				// 我需要监控这里
			}
		}
	}()
	//缓存预加载，缓存第一条数据
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		c.preCache(ctx, res)
	}()
	return res, nil
}

func (c *articleRepository) preCache(ctx context.Context, arts []domain.Article) {
	if len(arts) > 0 {
		err := c.cache.Set(ctx, arts[0])
		if err != nil {
			c.l.Error("提前预加载失败", logger.Error(err))
		}
	}
}

func (c *articleRepository) ToEntity(art domain.Article) dao.Article {
	return dao.Article{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,

		Status: uint8(art.Status),
	}
}

func (c *articleRepository) toDomain(art dao.Article) domain.Article {
	return domain.Article{
		Id:      art.Id,
		Title:   art.Title,
		Content: art.Content,
		Author: domain.Author{
			// 这里有一个错误
			Id: art.AuthorId,
		},
		Ctime:  time.UnixMilli(art.Ctime),
		Utime:  time.UnixMilli(art.Utime),
		Status: domain.ArticleStatus(art.Status),
	}
}

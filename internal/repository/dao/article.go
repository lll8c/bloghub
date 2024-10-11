package dao

import (
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type ArticleDAO interface {
	Insert(ctx context.Context, art Article) (int64, error)
	Update(ctx context.Context, art Article) error
	FindById(ctx context.Context, id int64) (Article, error)
	Sync(ctx context.Context, art Article) (int64, error)
	Upsert(ctx context.Context, art PublishArticle) error
	SyncStatus(ctx context.Context, artId int64, authorId int64, status int) error
	GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]Article, error)
	GetById(ctx context.Context, id int64) (Article, error)
	GetPubById(ctx context.Context, id int64) (PublishArticle, error)
	ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]PublishArticle, error)
}

type GROMArticleDAO struct {
	db *gorm.DB
}

func NewGROMArticleDAO(db *gorm.DB) ArticleDAO {
	return &GROMArticleDAO{
		db: db,
	}
}

// ListPub 获取7天内已发表的文章
func (a *GROMArticleDAO) ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]PublishArticle, error) {
	var res []PublishArticle
	const ArticleStatusPublished = 2
	err := a.db.WithContext(ctx).
		Where("utime < ? AND status = ?", start.UnixMilli(), ArticleStatusPublished).
		Order("utime DESC").Offset(offset).Limit(limit).
		Find(&res).Error
	return res, err
}

// SyncStatus 更新帖子状态
func (g *GROMArticleDAO) SyncStatus(ctx context.Context, artId int64, authorId int64, status int) error {
	now := time.Now().UnixMilli()
	return g.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		//更新制作库
		res := tx.Model(&Article{}).
			Where("id=? and author_id=?", artId, authorId).
			Updates(map[string]any{
				"status": status,
				"utime":  now,
			})
		if res.Error != nil {
			//数据库问题
			return res.Error
		}
		if res.RowsAffected != 1 {
			//要么Id错了，要么作者不对
			//用prometheus打点，只要频繁出现，就警告，然后手工介入排查
			return fmt.Errorf("可能有人修改其他人文章，id:%d, authorId:%d", artId, authorId)
		}
		//更新线上库
		return tx.Model(&Article{}).
			Where("id=?", artId).
			Updates(map[string]any{
				"status": status,
				"utime":  now,
			}).Error
	})
}

func (g *GROMArticleDAO) Sync(ctx context.Context, art Article) (int64, error) {
	//在事务内部，采用了闭包形式
	var (
		id  int64
		err error
	)
	err = g.db.Transaction(func(tx *gorm.DB) error {
		txDAO := NewGROMArticleDAO(tx)
		//操作制作库
		if art.Id > 0 {
			err = txDAO.Update(ctx, art)
		} else {
			id, err = txDAO.Insert(ctx, art)
		}
		if err != nil {
			return err
		}
		//操作线上库
		err = txDAO.Upsert(ctx, PublishArticle(art))
		return err
	})
	return id, err
}

// Upsert 插入或修改线上库
func (g *GROMArticleDAO) Upsert(ctx context.Context, art PublishArticle) error {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	//执行insert xxx on duplicate key update xxx
	return g.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"title":   art.Title,
			"content": art.Content,
			"status":  art.Status,
			"utime":   art.Utime,
		}),
	}).Create(&art).Error
	return nil
}

// Insert 插入制作库
func (g *GROMArticleDAO) Insert(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	err := g.db.WithContext(ctx).Create(&art).Error
	return art.Id, err
}

// Update 更新制作库
func (g *GROMArticleDAO) Update(ctx context.Context, art Article) error {
	now := time.Now().UnixMilli()
	art.Utime = now
	//直接指定要更新的具体字段
	res := g.db.WithContext(ctx).Model(&art).
		Where("id=? and author_id=?", art.Id, art.AuthorId).
		Updates(map[string]any{
			"title":   art.Title,
			"content": art.Content,
			"status":  art.Status,
			"utime":   art.Utime,
		})
	if res.Error != nil {
		return res.Error
	}
	//检查是否真的更新了，要返回一个err
	if res.RowsAffected == 0 {
		return errors.New("更新失败，可能是创作者非法")
	}
	return nil
}

func (g *GROMArticleDAO) FindById(ctx context.Context, id int64) (Article, error) {
	var article Article
	err := g.db.WithContext(ctx).Where("id=?", id).First(&article).Error
	return article, err
}

func (a *GROMArticleDAO) GetPubById(ctx context.Context, id int64) (PublishArticle, error) {
	var res PublishArticle
	err := a.db.WithContext(ctx).
		Where("id = ?", id).
		First(&res).Error
	return res, err
}

func (a *GROMArticleDAO) GetById(ctx context.Context, id int64) (Article, error) {
	var art Article
	err := a.db.WithContext(ctx).
		Where("id = ?", id).First(&art).Error
	return art, err
}

func (a *GROMArticleDAO) GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]Article, error) {
	var arts []Article
	err := a.db.WithContext(ctx).
		Where("author_id = ?", uid).
		Offset(offset).Limit(limit).
		// a ASC, B DESC
		Order("utime DESC").
		Find(&arts).Error
	return arts, err
}

type Article struct {
	Id    int64  `gorm:"primaryKey,autoIncrement" bson:"id,omitempty"`
	Title string `gorm:"type=varchar(4096)" bson:"title,omitempty"`
	//内容为大文本数据
	Content string `gorm:"type=BLOB" bson:"content,omitempty"`
	Status  uint8  `bson:"status,omitempty"`
	// 在作者id和创建时间上创建联合索引
	//AuthorId int64 `gorm:"index=aid_ctime"`
	//Ctime    int64 `gorm:"index=aid_ctime"`
	//在作者id上创建索引
	AuthorId int64 `gorm:"index" bson:"author_id,omitempty"`
	Ctime    int64 `bson:"ctime,omitempty"`
	Utime    int64 `bson:"utime,omitempty"`
}

type PublishArticle Article

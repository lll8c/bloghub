package dao

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ReaderDao interface {
	Insert(ctx context.Context, art Article) (int64, error)
	Update(ctx context.Context, art Article) error
	Upsert(ctx context.Context, art Article) error
}

type readerDao struct {
	db *gorm.DB
}

func newReaderDao(db *gorm.DB) ReaderDao {
	return &readerDao{db: db}
}

func (r *readerDao) Insert(ctx context.Context, art Article) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (r *readerDao) Update(ctx context.Context, art Article) error {
	//TODO implement me
	panic("implement me")
}

// Upsert 更新或者插入
func (r *readerDao) Upsert(ctx context.Context, art Article) error {
	//Id冲突，写不写都可以
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"title":   art.Title,
			"content": art.Content,
		}),
	}).Create(&art).Error

	return nil
}

package dao

import (
	"context"
	"gorm.io/gorm"
)

type AuthorDao interface {
	Insert(ctx context.Context, art Article) (int64, error)
	Update(ctx context.Context, art Article) error
}

type authorDao struct {
	db *gorm.DB
}

func newAuthorDao(db *gorm.DB) *authorDao {
	return &authorDao{db: db}
}

func (a authorDao) Insert(ctx context.Context, art Article) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (a authorDao) Update(ctx context.Context, art Article) error {
	//TODO implement me
	panic("implement me")
}

package repository

import (
	"context"
	"geektime/webook/internal/domain"
	"geektime/webook/internal/repository/cache"
	"geektime/webook/internal/repository/dao"
)

var (
	ErrUserDuplicateEmail = dao.ErrUserDuplicateEmail
	ErrUserNotFound       = dao.ErrUserNotFound
)

type UserRepository struct {
	dao   *dao.UserDAO
	cache *cache.UserCache
}

func NewUserRepository(dao *dao.UserDAO, cache *cache.UserCache) *UserRepository {
	return &UserRepository{
		dao:   dao,
		cache: cache,
	}
}

func (r *UserRepository) Create(ctx context.Context, u domain.User) error {
	return r.dao.Insert(ctx, dao.User{
		Email:    u.Email,
		Password: u.Password,
	})
}

func (r *UserRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
	//如果redis缓存里面找到就直接返回
	user, err := r.cache.Get(ctx, id)
	if err == nil {
		return user, err
	}
	//不管查找redis有没有出错，只要没找到，就查找数据库
	//后续采用限流或布隆过滤器防止数据库被冲垮
	ud, err := r.dao.FindById(ctx, id)
	u := domain.User{
		Id:       ud.Id,
		Email:    ud.Email,
		Password: ud.Password,
	}
	//从数据库中找到数据后，写到redis缓存中
	err = r.cache.Set(ctx, u)
	return u, err
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	u, err := r.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	//封装业务User
	return domain.User{
		Id:       u.Id,
		Email:    u.Email,
		Password: u.Password,
	}, nil
}

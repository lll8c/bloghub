package repository

import (
	"context"
	"database/sql"
	"geektime/webook/internal/domain"
	"geektime/webook/internal/repository/cache"
	"geektime/webook/internal/repository/dao"
	"time"
)

var (
	ErrUserDuplicate = dao.ErrUserDuplicate
	ErrUserNotFound  = dao.ErrUserNotFound
)

type UserRepository interface {
	Create(ctx context.Context, u domain.User) error
	FindById(ctx context.Context, id int64) (domain.User, error)
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	FindByPhone(ctx context.Context, phone string) (domain.User, error)
	FindByWechat(ctx context.Context, openId string) (domain.User, error)
	UpdateNonZeroFields(ctx context.Context, user domain.User) error
}

type CacheUserRepository struct {
	dao   dao.UserDAO
	cache cache.UserCache
}

func NewUserRepository(dao dao.UserDAO, cache cache.UserCache) UserRepository {
	return &CacheUserRepository{
		dao:   dao,
		cache: cache,
	}
}

func (r *CacheUserRepository) Create(ctx context.Context, u domain.User) error {
	return r.dao.Insert(ctx, domainToEntity(u))
}

func (r *CacheUserRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
	//如果redis缓存里面找到就直接返回
	user, err := r.cache.Get(ctx, id)
	if err == nil {
		return user, err
	}
	//不管查找redis有没有出错，只要没找到，就查找数据库
	//后续采用限流或布隆过滤器防止数据库被冲垮
	ud, err := r.dao.FindById(ctx, id)
	if err != nil {
		return domain.User{}, err
	}
	u := entityToDomain(ud)
	//从数据库中找到数据后，写到redis缓存中
	_ = r.cache.Set(ctx, u)
	return u, err
}

func (r *CacheUserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	u, err := r.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	//封装业务User
	return entityToDomain(u), nil
}

func (r *CacheUserRepository) FindByPhone(ctx context.Context, phone string) (domain.User, error) {
	u, err := r.dao.FindByPhone(ctx, phone)
	if err != nil {
		return domain.User{}, err
	}
	//封装业务User
	return entityToDomain(u), nil
}

func (r *CacheUserRepository) FindByWechat(ctx context.Context, openId string) (domain.User, error) {
	u, err := r.dao.FindByPhone(ctx, openId)
	if err != nil {
		return domain.User{}, err
	}
	//封装业务User
	return entityToDomain(u), nil
}

func (r *CacheUserRepository) UpdateNonZeroFields(ctx context.Context, user domain.User) error {
	return r.dao.UpdateById(ctx, domainToEntity(user))
}

func domainToEntity(u domain.User) dao.User {
	return dao.User{
		Id: u.Id,
		Email: sql.NullString{
			String: u.Email,
			Valid:  u.Email != "",
		},
		Phone: sql.NullString{
			String: u.Phone,
			Valid:  u.Phone != "",
		},
		WechatUnionId: sql.NullString{
			String: u.WechatInfo.UnionId,
			Valid:  u.WechatInfo.UnionId != "",
		},
		WechatOpenId: sql.NullString{
			String: u.WechatInfo.OpenId,
			Valid:  u.WechatInfo.OpenId != "",
		},
		Password: u.Password,
	}
}

func entityToDomain(u dao.User) domain.User {
	return domain.User{
		Id:       u.Id,
		Email:    u.Email.String,
		Password: u.Password,
		Ctime:    time.UnixMilli(u.Ctime),
		WechatInfo: domain.WechatInfo{
			OpenId:  u.WechatOpenId.String,
			UnionId: u.WechatUnionId.String,
		},
	}
}

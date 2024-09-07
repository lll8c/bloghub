package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"geektime/webook/internal/domain"
	"github.com/redis/go-redis/v9"
	"time"
)

var (
	ErrKeyNotExist = redis.Nil
)

// UserCache 面向接口编程
type UserCache struct {
	//传单机Redis可以
	//传cluster的Redis也可以
	client redis.Cmdable
	//过期时间
	expiration time.Duration
}

// NewUserCache
// A用到了B
// B一定接口
// B一定是A的字段
// A绝对不初始化B, 而是外面注入
func NewUserCache(client redis.Cmdable) *UserCache {
	return &UserCache{
		client:     client,
		expiration: time.Minute * 15,
	}
}

func (cache *UserCache) Get(ctx context.Context, id int64) (domain.User, error) {
	key := cache.key(id)
	val, err := cache.client.Get(ctx, key).Result()
	if err != nil {
		return domain.User{}, err
	}
	var u domain.User
	err = json.Unmarshal([]byte(val), &u)
	return u, err
}

func (cache *UserCache) Set(ctx context.Context, u domain.User) error {
	val, err := json.Marshal(u)
	if err != nil {
		return err
	}
	key := cache.key(u.Id)
	return cache.client.Set(ctx, key, val, cache.expiration).Err()
}

func (cache *UserCache) key(id int64) string {
	//user:info:123
	return fmt.Sprintf("user:info:%d", id)
}

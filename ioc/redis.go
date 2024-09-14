package ioc

import (
	"geektime/webook/config"
	"github.com/redis/go-redis/v9"
)

func InitRedis() redis.Cmdable {
	return redis.NewClient(&redis.Options{
		Addr: config.Config.Redis.Addr,
	})
}

func NewRatelimiter() redis.Limiter {
	return nil
}

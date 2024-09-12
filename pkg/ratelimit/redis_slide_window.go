package ratelimit

import (
	"context"
	_ "embed"
	"github.com/redis/go-redis/v9"
	"time"
)

//go:embed slide_window.lua
var luaScript string

// RedisSlidingWindowLimiter Redis上的滑动窗口算法限流器实现
type RedisSlidingWindowLimiter struct {
	cmd redis.Cmdable
	// 窗口大小
	interval time.Duration
	// 阈值
	rate int
}

func NewRedisSlidingWindowLimiter(cmd redis.Cmdable, interval time.Duration, rate int) *RedisSlidingWindowLimiter {
	return &RedisSlidingWindowLimiter{
		cmd:      cmd,
		interval: interval,
		rate:     rate,
	}
}

func (b *RedisSlidingWindowLimiter) Limit(ctx context.Context, key string) (bool, error) {
	//使用lua脚本解决并发问题
	return b.cmd.Eval(ctx, luaScript, []string{key},
		b.interval.Milliseconds(), b.rate, time.Now().UnixMilli()).Bool()
}

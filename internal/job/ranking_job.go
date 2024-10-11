package job

import (
	"context"
	"geektime/webook/internal/service"
	"geektime/webook/pkg/logger"
	rlock "github.com/gotomicro/redis-lock"
	"sync"
	"time"
)

type RankingJob struct {
	svc     service.RankingService
	l       logger.LoggerV1
	timeout time.Duration

	key       string
	client    *rlock.Client
	lock      *rlock.Lock
	localLock *sync.Mutex
}

func NewRankingJob(svc service.RankingService, l logger.LoggerV1,
	timeout time.Duration, client *rlock.Client) *RankingJob {
	return &RankingJob{
		svc:       svc,
		l:         l,
		timeout:   timeout,
		key:       "job:ranking",
		client:    client,
		localLock: &sync.Mutex{},
	}
}

func (r *RankingJob) Name() string {
	return "ranking"
}

func (r *RankingJob) Run() error {
	r.localLock.Lock()
	defer r.localLock.Unlock()
	//使用分布式锁，并且不会释放锁
	//r.lock为nil，表示这个节点没有拿到锁
	if r.lock == nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		//分布式锁，锁r.timeout时间
		lock, err := r.client.Lock(ctx, r.key, r.timeout, &rlock.FixIntervalRetry{
			Interval: time.Millisecond * 100,
			Max:      0,
		}, time.Second)
		if err != nil {
			//没拿到锁，可能其他节点拿到了锁
			return nil
		}
		r.lock = lock
		go func() {
			//自动续约机制，在一半的时候续约
			err = lock.AutoRefresh(r.timeout/2, time.Second)
			//续约失败
			if err != nil {
				r.localLock.Lock()
				r.lock = nil
				r.localLock.Unlock()
			}
		}()
	}

	// 这边就是你拿到了锁
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	return r.svc.TopN(ctx)
}

// Close 释放分布式锁
func (r *RankingJob) Close() error {
	r.localLock.Lock()
	lock := r.lock
	r.lock = nil
	r.localLock.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return lock.Unlock(ctx)
}

package scheduler

import (
	"context"
	"fmt"
	"geektime/webook/pkg/gormx/connpool"
	"geektime/webook/pkg/logger"
	"geektime/webook/pkg/migrator"
	"geektime/webook/pkg/migrator/events"
	"geektime/webook/pkg/migrator/validator"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"sync"
	"time"
)

// Scheduler 用来统一管理整个迁移过程
// 它不是必须的，你可以理解为这是为了方便用户操作（和你理解）而引入的。
type Scheduler[T migrator.Entity] struct {
	lock       sync.Mutex
	src        *gorm.DB
	dst        *gorm.DB
	pool       *connpool.DoubleWritePool
	l          logger.LoggerV1
	pattern    string
	cancelFull func()
	cancelIncr func()
	producer   events.Producer

	// 如果你要允许多个全量校验同时运行
	fulls map[string]func()
}

func NewScheduler[T migrator.Entity](
	l logger.LoggerV1,
	src *gorm.DB,
	dst *gorm.DB,
	// 这个是业务用的 DoubleWritePool
	pool *connpool.DoubleWritePool,
	producer events.Producer) *Scheduler[T] {
	return &Scheduler[T]{
		l:       l,
		src:     src,
		dst:     dst,
		pattern: connpool.PatternSrcOnly,
		cancelFull: func() {
			// 初始的时候，啥也不用做
		},
		cancelIncr: func() {
			// 初始的时候，啥也不用做
		},
		pool:     pool,
		producer: producer,
	}
}

// RegisterRoutes 这一个也不是必须的，就是你可以考虑利用配置中心，监听配置中心的变化
// 把全量校验，增量校验做成分布式任务，利用分布式任务调度平台来调度
func (s *Scheduler[T]) RegisterRoutes(server *gin.RouterGroup) {
	// 将这个暴露为 HTTP 接口
	// 你可以配上对应的 UI
	server.POST("/src_only", s.SrcOnly)
	server.POST("/src_first", s.SrcFirst)
	server.POST("/dst_first", s.DstFirst)
	server.POST("/dst_only", s.DstOnly)
	server.POST("/full/start", s.StartFullValidation)
	server.POST("/full/stop", s.StopFullValidation)
	server.POST("/incr/stop", s.StopIncrementValidation)
	server.POST("/incr/start", s.StartIncrementValidation)
}

// ---- 下面是四个阶段 ---- //

// SrcOnly 只读写源表
func (s *Scheduler[T]) SrcOnly(c *gin.Context) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.pattern = connpool.PatternSrcOnly
	s.pool.UpdatePattern(connpool.PatternSrcOnly)
	c.JSON(http.StatusOK, Result{
		Msg: "OK",
	})
}

func (s *Scheduler[T]) SrcFirst(c *gin.Context) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.pattern = connpool.PatternSrcFirst
	s.pool.UpdatePattern(connpool.PatternSrcFirst)
	c.JSON(http.StatusOK, Result{
		Msg: "OK",
	})
}

func (s *Scheduler[T]) DstFirst(c *gin.Context) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.pattern = connpool.PatternDstFirst
	s.pool.UpdatePattern(connpool.PatternDstFirst)
	c.JSON(http.StatusOK, Result{
		Msg: "OK",
	})
}

func (s *Scheduler[T]) DstOnly(c *gin.Context) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.pattern = connpool.PatternDstOnly
	s.pool.UpdatePattern(connpool.PatternDstOnly)
	c.JSON(http.StatusOK, Result{
		Msg: "OK",
	})
}

func (s *Scheduler[T]) StopIncrementValidation(c *gin.Context) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.cancelIncr()
	c.JSON(http.StatusOK, Result{
		Msg: "OK",
	})
}

func (s *Scheduler[T]) StartIncrementValidation(ctx *gin.Context) {
	var req StartIncrRequest
	if err := ctx.Bind(&req); err != nil {
		return
	}
	// 开启增量校验
	s.lock.Lock()
	defer s.lock.Unlock()
	// 取消上一次的
	cancel := s.cancelIncr
	v, err := s.newValidator()
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统异常",
		})
	}
	//增量校验相关参数
	v.Incr().Utime(req.Utime).
		SleepInterval(time.Duration(req.Interval) * time.Millisecond)

	go func() {
		var ctx context.Context
		ctx, s.cancelIncr = context.WithCancel(context.Background())
		cancel()
		err := v.Validate(ctx)
		s.l.Warn("退出增量校验", logger.Error(err))
	}()
	ctx.JSON(http.StatusOK, Result{
		Msg: "启动增量校验成功",
	})
}

func (s *Scheduler[T]) StopFullValidation(c *gin.Context) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.cancelFull()
	c.JSON(http.StatusOK, Result{
		Msg: "OK",
	})
}

// StartFullValidation 全量校验
func (s *Scheduler[T]) StartFullValidation(c *gin.Context) {
	// 可以考虑去重的问题
	s.lock.Lock()
	defer s.lock.Unlock()
	// 取消上一次的
	cancel := s.cancelFull
	v, err := s.newValidator()
	if err != nil {
		c.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统异常",
		})
	}
	var ctx context.Context
	ctx, s.cancelFull = context.WithCancel(context.Background())

	go func() {
		// 先取消上一次的
		cancel()
		err := v.Validate(ctx)
		if err != nil {
			s.l.Warn("退出全量校验", logger.Error(err))
		}
	}()
	c.JSON(http.StatusOK, Result{
		Msg: "OK",
	})
}

func (s *Scheduler[T]) newValidator() (*validator.Validator[T], error) {
	switch s.pattern {
	case connpool.PatternSrcOnly, connpool.PatternSrcFirst:
		return validator.NewValidator[T](s.src, s.dst, "SRC", s.l, s.producer), nil
	case connpool.PatternDstFirst, connpool.PatternDstOnly:
		return validator.NewValidator[T](s.dst, s.src, "DST", s.l, s.producer), nil
	default:
		return nil, fmt.Errorf("未知的 pattern %s", s.pattern)
	}
}

type StartIncrRequest struct {
	Utime int64 `json:"utime"`
	// 毫秒数
	// json 不能正确处理 time.Duration 类型
	Interval int64 `json:"interval"`
}

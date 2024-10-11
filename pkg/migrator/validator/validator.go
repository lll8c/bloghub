package validator

import (
	"context"
	"geektime/webook/pkg/logger"
	"geektime/webook/pkg/migrator"
	"geektime/webook/pkg/migrator/events"
	"github.com/ecodeclub/ekit/slice"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
	"time"
)

// Validator 数据校验
type Validator[T migrator.Entity] struct {
	//校验基准
	base *gorm.DB
	//校验目标
	target *gorm.DB

	l logger.LoggerV1
	//kafka生成者
	producer events.Producer
	//修复数据需要知道，是以src为准还是以dst为准
	direction string
	batchSize int
	//增量查询时间
	utime int64
	// <= 0 就认为中断
	// > 0 就认为睡眠
	sleepInterval time.Duration
	fromBase      func(ctx context.Context, offset int) (T, error)
}

func NewValidator[T migrator.Entity](
	base *gorm.DB,
	target *gorm.DB,
	direction string,
	l logger.LoggerV1,
	p events.Producer) *Validator[T] {
	res := &Validator[T]{
		base:   base,
		target: target,
		l:      l, producer: p,
		direction: direction,
		batchSize: 100}
	res.fromBase = res.fullFromBase
	return res
}

// Validate 校验者可以通过ctx控制校验程序退出
func (v *Validator[T]) Validate(ctx context.Context) error {
	//err := v.validateBaseToTarget(ctx)
	//if err != nil {
	//	return err
	//}
	//return v.validateTargetToBase(ctx)

	//使用goroutine正向和反向校验并行
	var eg errgroup.Group
	eg.Go(func() error {
		return v.validateBaseToTarget(ctx)
	})
	eg.Go(func() error {
		return v.validateTargetToBase(ctx)
	})
	return eg.Wait()
}

func (v *Validator[T]) validateBaseToTarget(ctx context.Context) error {
	offset := 0
	for {
		//一条一条的获取base数据库的数据
		src, err := v.fromBase(ctx, offset)
		if err == context.DeadlineExceeded || err == context.Canceled {
			return nil
		}
		if err == gorm.ErrRecordNotFound {
			// 你增量校验，要考虑一直运行的
			// 全量校验直接退出
			if v.sleepInterval <= 0 {
				return nil
			}
			//sleep后继续增量校验
			time.Sleep(v.sleepInterval)
			continue
		}
		if err != nil {
			// 查询出错了
			v.l.Error("base -> target 查询 base 失败", logger.Error(err))
			// 在这里，跳过这条数据的错误
			offset++
			continue
		}

		// 根据id获取目标数据库的数据
		var dst T
		err = v.target.WithContext(ctx).
			Where("id = ?", src.ID()).
			First(&dst).Error
		switch err {
		case gorm.ErrRecordNotFound:
			// target 没有
			// 丢一条消息到 Kafka 上
			v.notify(src.ID(), events.InconsistentEventTypeTargetMissing)
		case nil:
			//目标表找到了数据，开始比较
			equal := src.CompareTo(dst)
			//数据不一致
			if !equal {
				// 要丢一条消息到 Kafka 上
				v.notify(src.ID(), events.InconsistentEventTypeNEQ)
			}
		default:
			// 记录日志，然后继续
			// 做好监控
			v.l.Error("base -> target 查询 target 失败",
				logger.Int64("id", src.ID()),
				logger.Error(err))
		}
		offset++
	}
}

// 反向校验，查找base中被删除的数据
func (v *Validator[T]) validateTargetToBase(ctx context.Context) error {
	offset := 0
	for {
		//先批量获取目标数据库数据
		var ts []T
		err := v.target.WithContext(ctx).Select("id").
			//Where("utime > ?", v.utime).
			Order("id").Offset(offset).Limit(v.batchSize).
			Find(&ts).Error
		if err == context.DeadlineExceeded || err == context.Canceled {
			return nil
		}
		if err == gorm.ErrRecordNotFound || len(ts) == 0 {
			if v.sleepInterval <= 0 {
				return nil
			}
			time.Sleep(v.sleepInterval)
			continue
		}
		if err != nil {
			v.l.Error("target => base 查询 target 失败", logger.Error(err))
			offset += len(ts)
			continue
		}
		// 根据id集合查询base数据库
		var srcTs []T
		ids := slice.Map(ts, func(idx int, t T) int64 {
			return t.ID()
		})
		err = v.base.WithContext(ctx).Select("id").
			Where("id IN ?", ids).Find(&srcTs).Error
		if err == gorm.ErrRecordNotFound || len(srcTs) == 0 {
			// 都代表。base 里面一条对应的数据都没有
			v.notifyBaseMissing(ts)
			offset += len(ts)
			continue
		}
		if err != nil {
			v.l.Error("target => base 查询 base 失败", logger.Error(err))
			// 保守起见，我都认为 base 里面没有数据
			// v.notifyBaseMissing(ts)
			offset += len(ts)
			continue
		}
		// 找差集，diff 里面的，就是 target 有，但是 base 没有的
		diff := slice.DiffSetFunc(ts, srcTs, func(src, dst T) bool {
			return src.ID() == dst.ID()
		})
		v.notifyBaseMissing(diff)
		//没数据了
		if len(ts) < v.batchSize {
			if v.sleepInterval <= 0 {
				return nil
			}
			time.Sleep(v.sleepInterval)
		}
		offset += len(ts)
	}
}

func (v *Validator[T]) Utime(t int64) *Validator[T] {
	v.utime = t
	return v
}

func (v *Validator[T]) SleepInterval(interval time.Duration) *Validator[T] {
	v.sleepInterval = interval
	return v
}

func (v *Validator[T]) Full() *Validator[T] {
	v.fromBase = v.fullFromBase
	return v
}

func (v *Validator[T]) Incr() *Validator[T] {
	v.fromBase = v.incrFromBase
	return v
}

// 全量查找
func (v *Validator[T]) fullFromBase(ctx context.Context, offset int) (T, error) {
	dbCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	var src T
	err := v.base.WithContext(dbCtx).Order("id").
		Offset(offset).First(&src).Error
	return src, err
}

// 根据utime增量查找
func (v *Validator[T]) incrFromBase(ctx context.Context, offset int) (T, error) {
	dbCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	var src T
	err := v.base.WithContext(dbCtx).
		Where("utime > ?", v.utime).
		Order("utime").
		Offset(offset).First(&src).Error
	return src, err
}

func (v *Validator[T]) notifyBaseMissing(ts []T) {
	for _, val := range ts {
		v.notify(val.ID(), events.InconsistentEventTypeBaseMissing)
	}
}

func (v *Validator[T]) notify(id int64, typ string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	//生成者发送数据不一致消息
	err := v.producer.ProduceInconsistentEvent(ctx, events.InconsistentEvent{
		ID:        id,
		Type:      typ,
		Direction: v.direction,
	})
	if err != nil {
		v.l.Error("发送不一致消息失败",
			logger.Error(err),
			logger.String("type", typ),
			logger.Int64("id", id))
	}
}

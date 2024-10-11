package failover

import (
	"context"
	"geektime/webook/internal/service/sms"
	"sync/atomic"
)

// TimeoutFailoverSMSService
// 第二种策略
type TimeoutFailoverSMSService struct {
	//所有服务商
	svcs []sms.Service
	idx  int32
	//连续超时的个数
	cnt int32
	//阈值
	//连续超时超过这个数字，就切换服务商
	threshold int32
}

func NewTimeoutFailoverSMSService(svcs []sms.Service, cnt int32) *TimeoutFailoverSMSService {
	return &TimeoutFailoverSMSService{svcs: svcs, cnt: cnt}
}

func (t TimeoutFailoverSMSService) Send(ctx context.Context, tpl string, args []string, numbers ...string) error {
	idx := atomic.LoadInt32(&t.idx)
	cnt := atomic.LoadInt32(&t.cnt)
	//超时次数大于阈值，切换当前服务商
	if cnt > t.threshold {
		newIdx := (idx + 1) % int32(len(t.svcs))
		//并发问题
		//不能直接用atomic.AddInt32(&t.idx, 1)，不然一个线程切换一次
		//而是用CAS, 如果有一个执行了切换，那么后面线程就不用切换了
		if atomic.CompareAndSwapInt32(&t.idx, idx, newIdx) {
			//将cnt置0
			atomic.StoreInt32(&t.idx, 0)
		}
		// else 就是并发问题，其他线程已经将t.idx+1
		// 只要使用这个t.idx即可
		//idx = newIdx
		idx = atomic.LoadInt32(&t.idx)
	}

	svc := t.svcs[idx]
	err := svc.Send(ctx, tpl, args, numbers...)
	switch err {
	case nil:
		//正常发送了消息，连续超时次数置0
		atomic.StoreInt32(&t.cnt, 0)
		return nil
	case context.DeadlineExceeded, context.Canceled:
		//调用者超时，将超时次数cnt+1
		atomic.AddInt32(&t.cnt, 1)
		return err
	default:
		//其他错误
		return err
	}
}

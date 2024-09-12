package failover

import (
	"context"
	"errors"
	"geektime/webook/internal/service/sms"
	"sync/atomic"
)

// FailoverSMSService 自动切换服务商
// 第一种策略，直接轮询
type FailoverSMSService struct {
	//短信服务各种不同的实现，不同的服务商
	svcs []sms.Service
	//当前服务商下标
	idx uint64
}

func NewFailoverSMSService(svcs []sms.Service) *FailoverSMSService {
	return &FailoverSMSService{svcs: svcs}
}

func (f FailoverSMSService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	for _, svc := range f.svcs {
		err := svc.Send(ctx, tplId, args, numbers...)
		//发送成功
		if err == nil {
			return nil
		}
		//发送失败需要打印日志
		//这里需要监控
	}
	return errors.New("发送失败，所有服务商都尝试过了")
}

// SendV1 改进写法
func (f FailoverSMSService) SendV1(ctx context.Context, tplId string, args []string, numbers ...string) error {
	//我们取下一个节点作为起始节点，让负载均衡
	//注意使用原子操作
	idx := atomic.AddUint64(&f.idx, 1)
	length := uint64(len(f.svcs))
	//遍历所有服务商
	for i := idx; i < idx+length; i++ {
		svc := f.svcs[int(i%length)]
		err := svc.Send(ctx, tplId, args, numbers...)
		switch err {
		case nil:
			return nil
		case context.DeadlineExceeded, context.Canceled:
			//调用者超时或调用者主动取消了
			return err
		default:
			//输出日志
		}
	}
	return errors.New("发送失败，所有服务商都尝试过了")
}

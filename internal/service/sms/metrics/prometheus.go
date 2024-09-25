package metrics

import (
	"context"
	"geektime/webook/internal/service/sms"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

// PrometheusDecorator 使用prometheus记录短信服务耗时
type PrometheusDecorator struct {
	svc    sms.Service
	vector *prometheus.SummaryVec
}

func NewPrometheusDecorator(svc sms.Service) *PrometheusDecorator {
	vector := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: "lll",
		Subsystem: "webook",
		Name:      "sms_resp_time",
		Help:      "统计gin的http接口",
	}, []string{"tplId"})
	prometheus.MustRegister(vector)
	return &PrometheusDecorator{
		svc:    svc,
		vector: vector,
	}
}

func (s *PrometheusDecorator) Send(ctx context.Context, tpl string, args []string, numbers ...string) error {
	//记录耗时
	start := time.Now()
	defer func() {
		duration := time.Since(start).Milliseconds()
		s.vector.WithLabelValues(tpl).Observe(float64(duration))
	}()
	return s.svc.Send(ctx, tpl, args, numbers...)
}

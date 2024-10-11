package ioc

import (
	"geektime/webook/internal/job"
	"geektime/webook/internal/service"
	"geektime/webook/pkg/logger"
	rlock "github.com/gotomicro/redis-lock"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/robfig/cron/v3"
	"time"
)

func InitRankingJob(svc service.RankingService, l logger.LoggerV1,
	rlockClient *rlock.Client) *job.RankingJob {
	//一次运行不超过30秒
	return job.NewRankingJob(svc, l, time.Second*30, rlockClient)
}

func InitJobs(l logger.LoggerV1, rjob *job.RankingJob) *cron.Cron {
	builder := job.NewCronJobBuilder(l, prometheus.SummaryOpts{
		Namespace: "geekbang_daming",
		Subsystem: "webook",
		Name:      "cron_job",
		Help:      "定时任务执行",
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.75:  0.01,
			0.9:   0.01,
			0.99:  0.001,
			0.999: 0.0001,
		},
	})
	expr := cron.New(cron.WithSeconds())
	//定时执行job，每3分钟执行一次
	_, err := expr.AddJob("@every 3m", builder.Build(rjob))
	if err != nil {
		panic(err)
	}
	return expr
}

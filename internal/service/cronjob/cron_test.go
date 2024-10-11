package cronjob

import (
	cron "github.com/robfig/cron/v3"
	"log"
	"testing"
	"time"
)

func TestCronExpression(t *testing.T) {
	expr := cron.New(cron.WithSeconds())
	expr.AddJob("@every 1s", myJob{})
	expr.AddFunc("@every 1s", func() {
		t.Log("111")
	})
	expr.Start()
	//模拟运行了10s
	time.Sleep(time.Second * 10)
	//发出停止信号，但不会中断已经调度的任务
	stop := expr.Stop()
	//阻塞等待所有任务完成
	<-stop.Done()
}

type myJob struct {
}

func (m myJob) Run() {
	log.Println("222")
}

package events

import (
	"context"
	"geektime/webook/internal/domain"
	"geektime/webook/internal/events/article"
	"geektime/webook/internal/repository"
	"geektime/webook/pkg/logger"
	"geektime/webook/pkg/saramax"
	"github.com/IBM/sarama"
	"time"
)

type HistoryRecordConsumer struct {
	repo   repository.HistoryRecordRepository
	client sarama.Client
	l      logger.LoggerV1
}

func (i *HistoryRecordConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("interactive", i.client)
	if err != nil {
		return err
	}
	go func() {
		er := cg.Consume(context.Background(),
			[]string{"article_read"},
			saramax.NewHandler[article.ReadEvent](i.l, i.Consume))
		if er != nil {
			i.l.Error("退出消费", logger.Error(er))
		}
	}()
	return err
}

func (i *HistoryRecordConsumer) Consume(msg *sarama.ConsumerMessage,
	event article.ReadEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return i.repo.AddRecord(ctx, domain.HistoryRecord{
		BizId: event.Aid,
		Biz:   "article",
		Uid:   event.Uid,
	})
}

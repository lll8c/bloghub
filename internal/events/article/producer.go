package article

import (
	"context"
	"encoding/json"
	"github.com/IBM/sarama"
)

type Producer interface {
	ProduceReadEvent(ctx context.Context, evt ReadEvent) error
}

type KafkaProducer struct {
	producer sarama.SyncProducer
}

func NewKafkaProducer(producer sarama.SyncProducer) Producer {
	return &KafkaProducer{producer: producer}
}

// ProduceReadEvent 发送用户已读文章消息
func (k *KafkaProducer) ProduceReadEvent(ctx context.Context, evt ReadEvent) error {
	data, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	_, _, err = k.producer.SendMessage(&sarama.ProducerMessage{
		Topic: "article_read",
		Value: sarama.ByteEncoder(data),
	})
	return err
}

// ReadEvent 某个用户读了某篇文章
type ReadEvent struct {
	Uid int64
	Aid int64
}

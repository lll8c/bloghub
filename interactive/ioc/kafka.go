package ioc

import (
	events2 "geektime/webook/interactive/events"
	"geektime/webook/interactive/repository/dao"
	"geektime/webook/pkg/migrator/events/fixer"
	"geektime/webook/pkg/saramax"
	"github.com/IBM/sarama"
	"github.com/spf13/viper"
)

func InitKafkaClient() sarama.Client {
	type Config struct {
		Addr []string `yaml:"addr"`
	}
	var cfg Config
	err := viper.UnmarshalKey("kafka", &cfg)
	if err != nil {
		panic(err)
	}
	scfg := sarama.NewConfig()
	scfg.Producer.Return.Successes = true
	client, err := sarama.NewClient(cfg.Addr, scfg)
	if err != nil {
		panic(err)
	}
	return client
}

func InitSyncProducer(c sarama.Client) sarama.SyncProducer {
	p, err := sarama.NewSyncProducerFromClient(c)
	if err != nil {
		panic(err)
	}
	return p
}

func InitConsumers(
	c1 *events2.InteractiveReadEventConsumer, fixConsumer *fixer.Consumer[dao.Interactive]) []saramax.Consumer {
	return []saramax.Consumer{c1, fixConsumer}
}

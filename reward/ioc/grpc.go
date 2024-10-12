package ioc

import (
	"geektime/webook/pkg/grpcx"
	"geektime/webook/pkg/logger"
	grpc2 "geektime/webook/reward/grpc"
	"github.com/spf13/viper"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

func InitGRPCxServer(reward *grpc2.RewardServiceServer,
	ecli *clientv3.Client,
	l logger.LoggerV1) *grpcx.Server {
	type Config struct {
		Port     int    `yaml:"port"`
		EtcdAddr string `yaml:"etcdAddr"`
		EtcdTTL  int64  `yaml:"etcdTTL"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.server", &cfg)
	if err != nil {
		panic(err)
	}
	server := grpc.NewServer()
	reward.Register(server)
	return &grpcx.Server{
		Server: server,
		Port:   cfg.Port,
		Name:   "reward",
		L:      l,
		Client: ecli,
	}
}

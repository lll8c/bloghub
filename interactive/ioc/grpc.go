package ioc

import (
	grpc2 "geektime/webook/interactive/grpc"
	"geektime/webook/pkg/grpcx"
	"geektime/webook/pkg/logger"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func InitGRPCxServer(intrServer *grpc2.InteractiveServiceServer, l logger.LoggerV1) *grpcx.Server {
	type Config struct {
		EtcdAddrs []string `yaml:"etcdAddrs"`
		Port      int      `yaml:"port"`
		Name      string   `yaml:"name"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.server", &cfg)
	if err != nil {
		panic(err)
	}
	//先创建一个grpc的server
	server := grpc.NewServer()
	//注册intrServer
	intrServer.Register(server)
	return &grpcx.Server{
		Server:    server,
		Port:      cfg.Port,
		EtcdAddrs: cfg.EtcdAddrs,
		Name:      cfg.Name,
		L:         l,
	}
}

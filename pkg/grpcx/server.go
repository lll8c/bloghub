package grpcx

import (
	"context"
	"geektime/webook/pkg/logger"
	"geektime/webook/pkg/netx"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"google.golang.org/grpc"
	"net"
	"strconv"
	"time"
)

type Server struct {
	*grpc.Server
	Port      int
	EtcdAddrs []string
	Name      string
	L         logger.LoggerV1

	kaCancel func()
	Client   *etcdv3.Client
}

func (s *Server) ListenAndServe() error {
	l, err := net.Listen("tcp", ":"+strconv.Itoa(s.Port))
	if err != nil {
		return err
	}
	//嵌入服务注册过程
	s.Register()
	//这边类似gin.Run()
	return s.Server.Serve(l)
}

func (s *Server) Register() error {
	cli := s.Client
	em, err := endpoints.NewManager(cli, "service/"+s.Name)
	addr := netx.GetOutboundIP() + ":" + strconv.Itoa(s.Port)
	key := "service/" + s.Name + "/" + addr
	// 租期
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	//租期长度，单位：秒
	var ttl int64 = 10
	leaseResp, err := cli.Grant(ctx, ttl)
	//key是指这个实例的key
	//如果有instance id，用instance id, 如果没有用本机IP+端口
	err = em.AddEndpoint(ctx, key, endpoints.Endpoint{
		// 定位信息，客户端怎么连你
		Addr: addr,
	}, etcdv3.WithLease(leaseResp.ID))
	kaCtx, kaCancel := context.WithCancel(context.Background())
	s.kaCancel = kaCancel
	//进行续约
	ch, err := cli.KeepAlive(kaCtx, leaseResp.ID)
	if err != nil {
		return err
	}
	go func() {
		for kaResp := range ch {
			//大概ttl/3时候续约一次
			s.L.Debug(kaResp.String())
		}
	}()
	return nil
}

func (s *Server) Close() error {
	if s.kaCancel != nil {
		s.kaCancel()
	}
	if s.Client != nil {
		// 依赖注入，你就不要关
		return s.Client.Close()
	}
	s.GracefulStop()
	return nil
}

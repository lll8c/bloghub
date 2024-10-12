package main

import (
	"geektime/webook/pkg/ginx"
	"geektime/webook/pkg/grpcx"
	"geektime/webook/pkg/saramax"
)

type App struct {
	WebServer  *ginx.Server
	GRPCServer *grpcx.Server
	Consumers  []saramax.Consumer
}

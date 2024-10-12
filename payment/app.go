package main

import (
	"geektime/webook/pkg/ginx"
	"geektime/webook/pkg/grpcx"
)

type App struct {
	WebServer  *ginx.Server
	GRPCServer *grpcx.Server
}

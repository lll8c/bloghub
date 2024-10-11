package main

import (
	"geektime/webook/pkg/ginx"
	"geektime/webook/pkg/grpcx"
	"geektime/webook/pkg/saramax"
)

type App struct {
	server    *grpcx.Server
	consumers []saramax.Consumer
	webAdmin  *ginx.Server
}

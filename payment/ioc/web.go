package ioc

import (
	"geektime/webook/payment/web"
	"geektime/webook/pkg/ginx"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func InitGinServer(hdl *web.WechatHandler) *ginx.Server {
	engine := gin.Default()
	hdl.RegisterRoutes(engine)
	addr := viper.GetString("http.addr")
	/*	ginx.InitCounter(prometheus.CounterOpts{
		Namespace: "daming_geektime",
		Subsystem: "webook_payment",
		Name:      "http",
	})*/
	return &ginx.Server{
		Engine: engine,
		Addr:   addr,
	}
}

package main

import (
	"context"
	"fmt"
	"geektime/webook/ioc"
	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	"go.uber.org/zap"
	"net/http"
	"time"
)

func main() {
	//初始化viper，并读取配置
	initViper()
	app := InitApp()
	//初始化openTelemetry
	tpCancel := ioc.InitOTEL()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		tpCancel(ctx)
	}()
	//启动kafka消费者
	for _, c := range app.consumers {
		err := c.StartV1()
		if err != nil {
			panic(err)
		}
	}

	app.server.GET("/hello", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "hello")
	})
	app.server.Run("192.168.83.1:8080")
}

// 初始化prometheus
func initPrometheus() {
	go func() {
		// 专门给 prometheus 用的端口
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe("192.168.83.1:8081", nil)
	}()
}

func initViper() {
	viper.SetConfigFile("config/config.yaml")
	err := viper.ReadInConfig() // 查找并读取配置文件
	if err != nil {             // 处理读取配置文件的错误
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	viper.WatchConfig()
	viper.OnConfigChange(func(in fsnotify.Event) {
		err := viper.ReadInConfig() // 查找并读取配置文件
		if err != nil {             // 处理读取配置文件的错误
			panic(fmt.Errorf("Fatal error config file: %s \n", err))
		}
	})
}

// 要先启动etcd容器
func initViperRemote() {
	viper.SetConfigType("yaml")
	err := viper.AddRemoteProvider("etcd3",
		"127.0.0.1:12379", "/webook")
	if err != nil {
		panic(err)
	}
	err = viper.ReadRemoteConfig()
	if err != nil {
		panic(err)
	}
}

func initLogger() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(logger)
}

func initViperV1() {
	cfile := pflag.String("config", "config/config.yaml",
		"指定配置文件路径")
	pflag.Parse()
	viper.SetConfigFile(*cfile)
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}

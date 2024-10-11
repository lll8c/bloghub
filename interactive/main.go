package main

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

func main() {
	initViper()
	app := InitApp()
	//启动所有消费者
	for _, c := range app.consumers {
		err := c.Start()
		if err != nil {
			panic(err)
		}
	}
	//不停机数据迁移
	go func() {
		//后台管理端口
		err := app.webAdmin.Start()
		if err != nil {
			panic(err)
		}
	}()
	//监听，等待客户端连接
	err := app.server.ListenAndServe()
	if err != nil {
		panic(err)
	}
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

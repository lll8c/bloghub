package ioc

import (
	"fmt"
	"geektime/webook/internal/repository/dao"
	"geektime/webook/pkg/logger"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
	"time"
)

func InitDB(l logger.LoggerV1) *gorm.DB {
	//webook-mysql:11309
	//localhost:13316
	type Config struct {
		DSN string `yaml:"dsn"`
	}
	var cfg Config
	err := viper.UnmarshalKey("db", &cfg)
	if err != nil {
		panic(err)
	}
	fmt.Println(cfg.DSN)
	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{
		//日志
		Logger: glogger.New(gormLoggerFunc(l.Debug), glogger.Config{
			//慢查询阈值，只有执行时间超过这个阈值才会使用
			SlowThreshold: time.Millisecond * 10,
			LogLevel:      glogger.Info,
		}),
	})
	if err != nil {
		//初始化错误就直接panic
		panic(err)
	}
	//更新表结构
	err = dao.InitTables(db)
	if err != nil {
		panic(err)
	}
	return db
}

type gormLoggerFunc func(msg string, fields ...logger.Field)

func (g gormLoggerFunc) Printf(msg string, args ...interface{}) {
	g(msg, logger.Field{Key: "args", Val: args})
}

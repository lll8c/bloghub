package ioc

import (
	"fmt"
	dao2 "geektime/webook/interactive/repository/dao"
	"geektime/webook/internal/repository/dao"
	logger2 "geektime/webook/pkg/logger"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/opentelemetry/tracing"
	"gorm.io/plugin/prometheus"
)

func InitDB(l logger2.LoggerV1) *gorm.DB {
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
		/*Logger: glogger.New(gormLoggerFunc(l.Debug), glogger.Config{
			//慢查询阈值，只有执行时间超过这个阈值才会使用
			SlowThreshold: time.Second,
			LogLevel:      glogger.Info,
		}),*/
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		//初始化错误就直接panic
		panic(err)
	}

	//使用prometheus统计gorm
	err = db.Use(prometheus.New(prometheus.Config{
		DBName:          "webook",
		RefreshInterval: 15,
		StartServer:     false,
		MetricsCollector: []prometheus.MetricsCollector{
			&prometheus.MySQL{
				VariableNames: []string{"thread_running"},
			},
		},
	}))
	if err != nil {
		panic(err)
	}

	//使用OpenTelemetry统计gorm
	err = db.Use(tracing.NewPlugin(tracing.WithoutMetrics(),
		tracing.WithDBName("webook")))
	if err != nil {
		panic(err)
	}
	//不要记录metrics
	tracing.WithoutMetrics()
	//不要记录参数
	tracing.WithoutQueryVariables()

	//更新表结构
	err = dao.InitTables(db)
	if err != nil {
		panic(err)
	}
	err = dao2.InitTables(db)
	if err != nil {
		panic(err)
	}
	return db
}

type gormLoggerFunc func(msg string, fields ...logger2.Field)

func (g gormLoggerFunc) Printf(msg string, args ...interface{}) {
	g(msg, logger2.Field{Key: "args", Val: args})
}

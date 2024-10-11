package ioc

import (
	"geektime/webook/interactive/repository/dao"
	"geektime/webook/pkg/gormx/connpool"
	"geektime/webook/pkg/logger"
	logger2 "geektime/webook/pkg/logger"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type SrcDB *gorm.DB
type DstDB *gorm.DB

func InitSrcDB() SrcDB {
	return InitDB("src")
}

func InitDstDB() DstDB {
	return InitDB("dst")
}

func InitDoubleWritePool(src SrcDB, dst DstDB, l logger.LoggerV1) *connpool.DoubleWritePool {
	return connpool.NewDoubleWritePool(src, dst, l)
}

// InitBizDB 初始化双写数据库对象
func InitBizDB(p *connpool.DoubleWritePool) *gorm.DB {
	doubleWrite, err := gorm.Open(mysql.New(mysql.Config{
		Conn: p,
	}))
	if err != nil {
		panic(err)
	}
	return doubleWrite
}

func InitDB(key string) *gorm.DB {
	//webook-mysql:11309
	//localhost:13316
	type Config struct {
		DSN string `yaml:"dsn"`
	}
	var cfg Config
	err := viper.UnmarshalKey("db."+key, &cfg)
	if err != nil {
		panic(err)
	}
	//fmt.Println(cfg.DSN)
	db, err := gorm.Open(mysql.Open(cfg.DSN))
	if err != nil {
		//初始化错误就直接panic
		panic(err)
	}
	//更新表结构
	if err != nil {
		panic(err)
	}
	err = dao.InitTables(db)
	if err != nil {
		panic(err)
	}
	return db
}

type gormLoggerFunc func(msg string, fields ...logger2.Field)

func (g gormLoggerFunc) Printf(msg string, args ...interface{}) {
	g(msg, logger2.Field{Key: "args", Val: args})
}

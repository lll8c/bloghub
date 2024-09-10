package ioc

import (
	"geektime/webook/config"
	"geektime/webook/internal/repository/dao"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitDB() *gorm.DB {
	//webook-mysql:11309
	//localhost:13316
	db, err := gorm.Open(mysql.Open(config.Config.DB.DSN))
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

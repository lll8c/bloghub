package main

import (
	"geektime/webook/interactive/repository/dao"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"testing"
)

func TestDB(t *testing.T) {
	db, err := gorm.Open(mysql.Open("root:root@tcp(localhost:13316)/webook"))
	if err != nil {
		//初始化错误就直接panic
		panic(err)
	}
	var i int64
	for i = 1; i < 10; i++ {
		db.Create(&dao.Interactive{
			BizId:      i,
			Biz:        "test",
			ReadCnt:    i + 1,
			LikeCnt:    i + 2,
			CollectCnt: i + 3,
			Utime:      0,
			Ctime:      0,
		})
	}
}

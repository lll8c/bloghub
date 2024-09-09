package main

import (
	"geektime/webook/config"
	"geektime/webook/internal/repository"
	"geektime/webook/internal/repository/cache"
	"geektime/webook/internal/repository/dao"
	"geektime/webook/internal/service"
	"geektime/webook/internal/service/sms/localsms"
	"geektime/webook/internal/web"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"net/http"
)

func main() {
	r := web.InitWeb()
	db := initDB()
	client := InitRedis()
	u := initUser(db, client)
	u.RegisterRoutes(r)

	//r := gin.Default()
	r.GET("/hello", func(c *gin.Context) {
		c.String(http.StatusOK, "hello, nihao")
	})
	r.Run("0.0.0.0:8080")
}

func initUser(db *gorm.DB, client redis.Cmdable) *web.UserHandler {
	dao := dao.NewUserDao(db)
	userCache := cache.NewUserCache(client)
	repo := repository.NewUserRepository(dao, userCache)
	svc := service.NewUserService(repo)

	codeCache := cache.NewCodeCache(client)
	codeRepo := repository.NewCodeRepository(codeCache)
	//暂时不用真的发短信服务
	/*codeClient, err := dysmsapi.NewClientWithAccessKey("cn-hunan",
		"...",
		"...")
	if err != nil {
		panic(err)
	}
	smsSvc := aliyun.NewService("小微书", codeClient)*/
	//模拟一下发短信服务，方便测试
	smsSvc := localsms.NewService()
	codeSvc := service.NewCodeService(codeRepo, smsSvc, "SMS_472665076")
	u := web.NewUserHandler(svc, codeSvc)
	return u
}

func initDB() *gorm.DB {
	//webook-mysql:11309
	//localhost:13316
	db, err := gorm.Open(mysql.Open(config.Config.DB.DSN))
	if err != nil {
		//初始化错误就直接panic
		panic(err)
	}
	err = dao.InitTables(db)
	if err != nil {
		panic(err)
	}
	return db
}

func InitRedis() redis.Cmdable {
	return redis.NewClient(&redis.Options{
		Addr: config.Config.Redis.Addr,
	})
}

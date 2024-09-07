package main

import (
	"geektime/webook/config"
	"geektime/webook/internal/repository"
	"geektime/webook/internal/repository/dao"
	"geektime/webook/internal/service"
	"geektime/webook/internal/web"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"net/http"
)

func main() {
	r := web.RegisterRoutes()
	db := initDB()
	u := initUser(db)
	_ = redis.NewClient(&redis.Options{
		Addr: config.Config.Redis.Addr,
	})
	r.POST("users/signup", u.SignUp)
	r.POST("users/login", u.Login)
	r.POST("users/edit", u.Edit)
	r.POST("users/profile", u.Profile)

	//r := gin.Default()
	r.GET("/hello", func(c *gin.Context) {
		c.String(http.StatusOK, "hello, nihao")
	})
	r.Run("0.0.0.0:8080")
}

func initUser(db *gorm.DB) *web.UserHandler {
	dao := dao.NewUserDao(db)
	repo := repository.NewUserRepository(dao)
	svc := service.NewUserService(repo)
	u := web.NewUserHandler(svc)
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

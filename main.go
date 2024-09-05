package main

import (
	"geektime/webook/internal/repository"
	"geektime/webook/internal/repository/dao"
	"geektime/webook/internal/service"
	"geektime/webook/internal/web"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"net/http"
)

func main() {
	//r := web.RegisterRoutes()
	//db := initDB()
	//u := initUser(db)

	//r.POST("users/signup", u.SignUp)
	//r.POST("users/login", u.Login)
	//r.POST("users/edit", u.Edit)
	//r.POST("users/profile", u.Profile)
	r := gin.Default()
	r.GET("/hello", func(c *gin.Context) {
		c.String(http.StatusOK, "hello")
	})
	r.Run(":8080")
}

func initUser(db *gorm.DB) *web.UserHandler {
	dao := dao.NewUserDao(db)
	repo := repository.NewUserRepository(dao)
	svc := service.NewUserService(repo)
	u := web.NewUserHandler(svc)
	return u
}

func initDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open("root:root@tcp(localhost:13316)/webook"))
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

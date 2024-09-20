package pkg

import (
	"geektime/webook/pkg/logger"
	"github.com/gin-gonic/gin"
	"net/http"
)

//这个东西，放到自己的ginx插件亏中
//技术含量不高，但有技巧

func WrapBody[T any](l logger.LoggerV1, fn func(ctx *gin.Context, req T) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req T
		if err := ctx.Bind(&req); err != nil {
			return
		}
		//业务逻辑
		res, err := fn(ctx, req)
		if err != nil {
			//处理error
			l.Error("出错", logger.Error(err))
		}
		ctx.JSON(http.StatusOK, res)
	}
}

type Result struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

package web

import "github.com/gin-gonic/gin"

type ArticleHandler struct {
}

func (h *ArticleHandler) RegisterRoutes(r *gin.Engine) {
	g := r.Group("/articles")

	g.POST("/edit", h.Edit)
	g.POST("/publish", h.Publish)
	g.POST("/withdraw", h.Withdraw)
}

func (h *ArticleHandler) Edit(context *gin.Context) {

}

func (h *ArticleHandler) Publish(context *gin.Context) {

}

func (h *ArticleHandler) Withdraw(context *gin.Context) {

}

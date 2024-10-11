package web

import (
	intrv1 "geektime/webook/api/proto/gen/intr/v1"
	"geektime/webook/internal/domain"
	"geektime/webook/internal/service"
	jwt2 "geektime/webook/internal/web/jwt"
	"geektime/webook/pkg/logger"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"time"
)

type ArticleHandler struct {
	svc     service.ArticleService
	l       logger.LoggerV1
	intrSvc intrv1.InteractiveServiceClient
	biz     string
}

func NewArticleHandler(articleSvc service.ArticleService, l logger.LoggerV1, intrSvc intrv1.InteractiveServiceClient) *ArticleHandler {
	return &ArticleHandler{
		svc:     articleSvc,
		l:       l,
		intrSvc: intrSvc,
		biz:     "article",
	}
}

func (h *ArticleHandler) RegisterRoutes(r *gin.Engine) {
	g := r.Group("/articles")

	g.POST("/edit", h.Edit)
	g.POST("/publish", h.Publish)
	g.POST("/withdraw", h.Withdraw)
	// 创作者接口
	// 按照道理来说，这边就是 GET 方法 /list?offset=?&limit=?
	g.POST("/list", h.List)
	g.GET("/detail/:id", h.Detail)

	pub := g.Group("/pub")
	pub.GET("/:id", h.PubDetail)
	//点赞或取消点赞
	pub.POST("/like", h.Like)
	pub.POST("/collect", h.Collect)
}

func (h *ArticleHandler) Edit(ctx *gin.Context) {
	type Req struct {
		Id      int64  `json:"id"`
		Title   string `json:"title,omitempty"`
		Content string `json:"content,omitempty"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	c := ctx.MustGet("claims")
	claims, ok := c.(*jwt2.UserClaims)
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("未发现用户的session信息")
		return
	}
	id, err := h.svc.Save(ctx, domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: claims.Uid,
		},
	})
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("保存帖子失败", logger.Error(err))
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg:  "Ok",
		Data: id,
	})
}

func (h *ArticleHandler) Publish(ctx *gin.Context) {
	type Req struct {
		Id      int64  `json:"id"`
		Title   string `json:"title,omitempty"`
		Content string `json:"content,omitempty"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	c := ctx.MustGet("claims")
	claims, ok := c.(*jwt2.UserClaims)
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("未发现用户的session信息")
		return
	}
	id, err := h.svc.Publish(ctx, domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: claims.Uid,
		},
	})
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("发表帖子失败", logger.Error(err))
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg:  "Ok",
		Data: id,
	})

}

// Withdraw 设置帖子不可见
func (h *ArticleHandler) Withdraw(ctx *gin.Context) {
	type Req struct {
		Id int64 `json:"id"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	c := ctx.MustGet("claims")
	claims, ok := c.(*jwt2.UserClaims)
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("未发现用户的session信息")
		return
	}
	err := h.svc.Withdraw(ctx, domain.Article{
		Id: req.Id,
		Author: domain.Author{
			Id: claims.Uid,
		},
	})
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("更新帖子状态失败", logger.Error(err))
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "Ok",
	})
}

func (h *ArticleHandler) List(ctx *gin.Context) {
	var page Page
	if err := ctx.Bind(&page); err != nil {
		return
	}
	// 我要不要检测一下？
	uc := ctx.MustGet("claims").(*jwt2.UserClaims)
	arts, err := h.svc.GetByAuthor(ctx, uc.Uid, page.Offset, page.Limit)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("查找文章列表失败",
			logger.Error(err),
			logger.Int("offset", page.Offset),
			logger.Int("limit", page.Limit),
			logger.Int64("uid", uc.Uid))
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Data: slice.Map[domain.Article, ArticleVo](arts, func(idx int, src domain.Article) ArticleVo {
			return ArticleVo{
				Id:    src.Id,
				Title: src.Title,
				//摘要
				Abstract: src.Abstract(),
				//Content:  src.Content,
				AuthorId: src.Author.Id,
				// 列表，你不需要
				Status: src.Status.ToUint8(),
				Ctime:  src.Ctime.Format(time.DateTime),
				Utime:  src.Utime.Format(time.DateTime),
			}
		}),
	})
}

func (h *ArticleHandler) Detail(ctx *gin.Context) {
	idstr := ctx.Param("id")
	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg:  "id 参数错误",
			Code: 4,
		})
		h.l.Warn("查询文章失败，id 格式不对",
			logger.String("id", idstr),
			logger.Error(err))
		return
	}
	art, err := h.svc.GetById(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg:  "系统错误",
			Code: 5,
		})
		h.l.Error("查询文章失败",
			logger.Int64("id", id),
			logger.Error(err))
		return
	}
	uc := ctx.MustGet("claims").(*jwt2.UserClaims)
	if art.Author.Id != uc.Uid {
		// 有人在搞鬼
		ctx.JSON(http.StatusOK, Result{
			Msg:  "系统错误",
			Code: 5,
		})
		h.l.Error("非法查询文章",
			logger.Int64("id", id),
			logger.Int64("uid", uc.Uid))
		return
	}

	vo := ArticleVo{
		Id:    art.Id,
		Title: art.Title,
		//Abstract: art.Abstract(),

		Content:  art.Content,
		AuthorId: art.Author.Id,
		// 列表，你不需要
		Status: art.Status.ToUint8(),
		Ctime:  art.Ctime.Format(time.DateTime),
		Utime:  art.Utime.Format(time.DateTime),
	}
	ctx.JSON(http.StatusOK, Result{Data: vo})
}

func (h *ArticleHandler) PubDetail(ctx *gin.Context) {
	idstr := ctx.Param("id")
	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg:  "id 参数错误",
			Code: 4,
		})
		h.l.Warn("查询文章失败，id 格式不对",
			logger.String("id", idstr),
			logger.Error(err))
		return
	}
	uc := ctx.MustGet("claims").(*jwt2.UserClaims)
	art, err := h.svc.GetPubById(ctx, id, uc.Uid)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg:  "系统错误",
			Code: 5,
		})
		h.l.Error("查询文章失败，系统错误",
			logger.Error(err),
			logger.Int64("id", id),
			logger.Int64("uid", uc.Uid))
		return
	}

	//获取这篇文章的所有计数
	resp, err := h.intrSvc.Get(ctx, &intrv1.GetRequest{
		Biz:   h.biz,
		BizId: id,
		Uid:   uc.Uid,
	})
	if err != nil {
		//这个地方可以容忍错误
		/*ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})*/
		h.l.Error("获取文章信息失败", logger.Error(err))
		//return
	}
	//已经通过kafka实现
	/*//增加阅读计数
	//开goroutine异步执行
	go func() {
		//不要复用err
		er := h.intrSvc.IncrReadCnt(ctx, h.biz, id)
		if er != nil {
			h.l.Error("增加阅读计数失败",
				logger.Int64("aid", id),
				logger.Error(err))
		}
	}()*/
	intr := resp.Intr
	ctx.JSON(http.StatusOK, Result{
		Data: ArticleVo{
			Id:    art.Id,
			Title: art.Title,

			Content:    art.Content,
			AuthorId:   art.Author.Id,
			AuthorName: art.Author.Name,

			Status: art.Status.ToUint8(),

			ReadCnt:    intr.ReadCnt,
			LikeCnt:    intr.LikeCnt,
			CollectCnt: intr.CollectCnt,
			Liked:      intr.Liked,
			Collected:  intr.Collected,

			Ctime: art.Ctime.Format(time.DateTime),
			Utime: art.Utime.Format(time.DateTime),
		},
	})
}

// Like 点赞或取消点赞
func (h *ArticleHandler) Like(c *gin.Context) {
	type Req struct {
		Id int64 `json:"id"`
		// true 是点赞，false 是不点赞
		Like bool `json:"like"`
	}
	var req Req
	if err := c.Bind(&req); err != nil {
		return
	}
	uc := c.MustGet("claims").(*jwt2.UserClaims)
	var err error
	if req.Like {
		// 点赞
		_, err = h.intrSvc.Like(c, &intrv1.LikeRequest{
			Biz:   h.biz,
			BizId: req.Id,
			Uid:   uc.Uid,
		})
	} else {
		// 取消点赞
		_, err = h.intrSvc.CancelLike(c, &intrv1.CancelLikeRequest{
			Biz:   h.biz,
			BizId: req.Id,
			Uid:   uc.Uid,
		})
	}
	if err != nil {
		c.JSON(http.StatusOK, Result{
			Code: 5, Msg: "系统错误",
		})
		h.l.Error("点赞/取消点赞失败",
			logger.Error(err),
			logger.Int64("uid", uc.Uid),
			logger.Int64("aid", req.Id))
		return
	}
	c.JSON(http.StatusOK, Result{
		Msg: "OK",
	})
}

func (h *ArticleHandler) Collect(ctx *gin.Context) {
	type Req struct {
		Id  int64 `json:"id"`
		Cid int64 `json:"cid"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	uc := ctx.MustGet("claims").(*jwt2.UserClaims)

	_, err := h.intrSvc.Collect(ctx, &intrv1.CollectRequest{
		Biz:   h.biz,
		BizId: req.Id,
		Cid:   req.Cid,
		Uid:   uc.Uid,
	})
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5, Msg: "系统错误",
		})
		h.l.Error("收藏失败",
			logger.Error(err),
			logger.Int64("uid", uc.Uid),
			logger.Int64("aid", req.Id))
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "OK",
	})
}

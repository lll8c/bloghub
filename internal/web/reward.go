package web

import (
	rewardv1 "geektime/webook/api/proto/gen/reward/v1"
	"geektime/webook/internal/service"
	"geektime/webook/internal/web/jwt"
	"github.com/gin-gonic/gin"
	"net/http"
)

type RewardHandler struct {
	articleSvc   service.ArticleService
	rewardClient rewardv1.RewardServiceClient
}

func NewRewardHandler(rewardClient rewardv1.RewardServiceClient, articleSvc service.ArticleService) *RewardHandler {
	return &RewardHandler{
		rewardClient: rewardClient,
		articleSvc:   articleSvc}
}

func (h *RewardHandler) RegisterRoutes(server *gin.Engine) {
	rg := server.Group("/reward")
	rg.POST("/detail", h.GetReward)
}

// Reward 通过文章id打赏作者
func (h *RewardHandler) Reward(ctx *gin.Context) {
	type ArticleRewardReq struct {
		Id  int64 `json:"id"`
		Amt int64 `json:"amt"`
	}
	var req ArticleRewardReq
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "参数绑定错误",
		})
	}
	uc, _ := ctx.MustGet("claims").(*jwt.UserClaims)

	// 在这里分发
	// h.reward.WechatPreReward
	// h.reward.AlipayPreReward
	artResp, err := h.articleSvc.GetPubById(ctx, req.Id, uc.Uid)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
	}
	// 最关键的一步骤，就是拿到二维码
	resp, err := h.rewardClient.PreReward(ctx, &rewardv1.PreRewardRequest{
		Biz:       "article",
		BizId:     artResp.Id,
		BizName:   artResp.Title,
		TargetUid: artResp.Author.Id,
		Uid:       uc.Uid,
		Amt:       req.Amt,
	})
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
	}
	ctx.JSON(http.StatusOK, Result{
		Data: map[string]any{
			"codeURL": resp.CodeUrl,
			"rid":     resp.Rid,
		},
	})
}

// GetReward 通过id查询自己的打赏记录
func (h *RewardHandler) GetReward(ctx *gin.Context) {
	type GetRewardReq struct {
		Rid int64
	}
	var req GetRewardReq
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "参数绑定错误",
		})
	}
	claims, _ := ctx.MustGet("claims").(*jwt.UserClaims)
	resp, err := h.rewardClient.GetReward(ctx, &rewardv1.GetRewardRequest{
		// 我这一次打赏的 ID
		Rid: req.Rid,
		// 要防止非法访问，我只能看到我打赏的记录
		// 我不能看到别人打赏记录
		Uid: claims.Uid,
	})
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
	}
	ctx.JSON(http.StatusOK, Result{
		// 暂时也就是只需要状态
		Data: resp.Status.String(),
	})
}

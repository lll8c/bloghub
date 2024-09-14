package web

import (
	"fmt"
	"geektime/webook/internal/service"
	"geektime/webook/internal/service/oauth/wechat"
	jwt2 "geektime/webook/internal/web/jwt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	uuid "github.com/lithammer/shortuuid/v4"
	"net/http"
	"time"
)

type OAuth2WechatHandler struct {
	svc     wechat.Service
	userSvc service.UserService
	cfg     Config
	jwt2.JwtHandler
}

type Config struct {
	Secure   bool
	StateKey string
}

func NewOAuth2WechatHandler(svc wechat.Service, userSvc service.UserService, jwtHdl jwt2.JwtHandler) *OAuth2WechatHandler {
	return &OAuth2WechatHandler{
		svc:        svc,
		userSvc:    userSvc,
		JwtHandler: jwtHdl,
		cfg: Config{
			Secure:   false,
			StateKey: "secret",
		},
	}
}

func (h *OAuth2WechatHandler) RegisterRoutes(r *gin.Engine) {
	g := r.Group("/oauth2/wechat")
	g.GET("/authurl", h.AuthURL)
	g.Any("/callback", h.Callback)
}

func (h *OAuth2WechatHandler) AuthURL(ctx *gin.Context) {
	//生成state参数
	state := uuid.New()
	val, err := h.svc.AuthURL(ctx, state)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg:  "构造跳转URL失败",
			Code: 5,
		})
		return
	}
	if err := h.setStateCookie(ctx, state); err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg:  "设置state错误",
			Code: 5,
		})
		return
	}
	//将微信地址返回给前端
	ctx.JSON(http.StatusOK, Result{
		Data: val,
	})
}

func (h *OAuth2WechatHandler) setStateCookie(ctx *gin.Context, state string) error {
	//使用token存state
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, StateClaims{
		State: state,
		RegisteredClaims: jwt.RegisteredClaims{
			//过期时间为预期扫码登录的时间
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
		},
	})
	tokenStr, err := token.SignedString([]byte(h.cfg.StateKey))
	if err != nil {
		return err
	}
	//将stateToken存到cookie中
	ctx.SetCookie("jwt-state", tokenStr, 600, "/oauth2/wechat/callback/",
		"", h.cfg.Secure, true)
	return nil
}

// Callback 扫码后的回调方法
func (h *OAuth2WechatHandler) Callback(ctx *gin.Context) {
	code := ctx.Query("code")
	state := ctx.Query("state")
	//校验浏览器的state和微信回调返回的state
	err := h.VerifyState(ctx, state)
	if err != nil {
		if err != nil {
			ctx.JSON(http.StatusOK, Result{
				Code: 5,
				Msg:  "系统错误",
			})
			return
		}
	}
	//向微信请求 获取授权码和ID
	info, err := h.svc.VerifyCode(ctx, code)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	//获取到了授权码，实现登录
	//userService 通过获取的openId查找或创建user
	user, err := h.userSvc.FindOrCreateByWechat(ctx, info.OpenId)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	//用jwt设置登录状态
	//创建并设置token
	if err := h.SetLoginToken(ctx, user.Id); err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	//登录成功
	ctx.JSON(http.StatusOK, Result{
		Msg: "OK",
	})
}

func (h *OAuth2WechatHandler) VerifyState(ctx *gin.Context, state string) error {
	//校验state
	//获取stateTokenStr
	ck, err := ctx.Cookie("jwt-state")
	if err != nil {
		//被伪造了请求
		return fmt.Errorf("拿不到state的cookie %w", err)
	}
	var stateClaim StateClaims
	token, err := jwt.ParseWithClaims(ck, &stateClaim, func(token *jwt.Token) (interface{}, error) {
		return h.cfg.StateKey, nil
	})
	if err != nil || !token.Valid {
		return fmt.Errorf("token已经过期了 %w", err)
	}
	if stateClaim.State != state {
		return fmt.Errorf("state不相等")
	}
	return nil
}

type StateClaims struct {
	jwt.RegisteredClaims
	State string
}

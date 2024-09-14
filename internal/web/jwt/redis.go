package jwt

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
)

type RedisJWTHandler struct {
	signingMethod jwt.SigningMethod
	//长token过期时间
	rcExpiration time.Duration
	cmd          redis.Cmdable
}

func NewRedisJWTHandler(cmd redis.Cmdable) JwtHandler {
	return &RedisJWTHandler{
		signingMethod: jwt.SigningMethodHS512,
		rcExpiration:  time.Hour * 24 * 7,
		cmd:           cmd,
	}
}

// ExtractToken 提取token
func (h *RedisJWTHandler) ExtractToken(ctx *gin.Context) string {
	//Bearer xxx.xxx.xxx
	tokenHeader := ctx.GetHeader("Authorization")
	segs := strings.Split(tokenHeader, " ")
	if len(segs) != 2 {
		//格式不对，没登录
		return ""
	}
	return segs[1]
}

func (h *RedisJWTHandler) SetLoginToken(ctx *gin.Context, uid int64) error {
	//使用ssid标识用户登录状态
	ssid := uuid.New().String()
	err := h.SetJWTToken(ctx, uid, ssid)
	if err != nil {
		return err
	}
	err = h.SetRefreshToken(ctx, uid, ssid)
	return err
}

// SetJWTToken 设置短token
func (h *RedisJWTHandler) SetJWTToken(ctx *gin.Context, uid int64, ssid string) error {
	claims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			//设置过期时间
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
		},
		UserAgent: ctx.Request.UserAgent(),
		Uid:       uid,
		Ssid:      ssid,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte("secret"))
	if err != nil {
		return err
	}
	//将token放到header中
	ctx.Header("x-jwt-token", tokenStr)
	return nil
}

// SetRefreshToken 设置长token
func (h *RedisJWTHandler) SetRefreshToken(ctx *gin.Context, uid int64, ssid string) error {
	rc := RefreshClaims{
		Uid:  uid,
		Ssid: ssid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(h.rcExpiration)),
		},
	}
	token := jwt.NewWithClaims(h.signingMethod, rc)
	tokenStr, err := token.SignedString(RCJWTKey)
	if err != nil {
		return err
	}
	ctx.Header("x-refresh-token", tokenStr)
	return nil
}

// CheckSession 从redis中查询标识，判断是否已经退出登录
func (h *RedisJWTHandler) CheckSession(ctx *gin.Context, ssid string) error {
	cnt, err := h.cmd.Exists(ctx, fmt.Sprintf("users:ssid:%s", ssid)).Result()
	if err != nil {
		return err
	}
	if cnt > 0 {
		return errors.New("token 无效")
	}
	return nil
}

// ClearToken 清理token，退出登录
func (h *RedisJWTHandler) ClearToken(ctx *gin.Context) error {
	ctx.Header("x-jwt-token", "")
	ctx.Header("x-refresh-token", "")
	claims := ctx.MustGet("claims").(*UserClaims)
	err := h.cmd.Set(ctx, fmt.Sprintf("users:ssid:%s", claims.Ssid), "", h.rcExpiration).Err()
	return err
}

var JWTKey = []byte("k6CswdUm77WKcbM68UQUuxVsHSpTCwgK")
var RCJWTKey = []byte("k6CswdUm77WKcbM68UQUuxVsHSpTCwgA")

type RefreshClaims struct {
	jwt.RegisteredClaims
	Uid  int64
	Ssid string
}

type UserClaims struct {
	jwt.RegisteredClaims
	//声明你自己的要放进去token里面的数据
	Uid       int64
	Ssid      string
	UserAgent string
}

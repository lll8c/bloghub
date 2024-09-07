package web

import (
	"fmt"
	"geektime/webook/internal/domain"
	"geektime/webook/internal/service"
	"geektime/webook/internal/web/middleware"
	"github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"time"
)

// UserHandler 定义与用户相关的路由
type UserHandler struct {
	emailExp    *regexp2.Regexp //邮箱校验器
	passwordExp *regexp2.Regexp //密码校验器
	svc         *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	//用于校验密码和邮箱的正则表达式
	const (
		emailRegexPattern = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
		// 和上面比起来，用 ` 看起来就比较清爽
		passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
	)
	//初始化校验器
	return &UserHandler{
		svc:         svc,
		emailExp:    regexp2.MustCompile(emailRegexPattern, regexp2.None),
		passwordExp: regexp2.MustCompile(passwordRegexPattern, regexp2.None),
	}
}

func (u *UserHandler) SignUp(ctx *gin.Context) {
	type SignUpReq struct {
		Email           string `json:"email,omitempty"`
		Password        string `json:"password,omitempty"`
		ConfirmPassword string `json:"confirm_password,omitempty"`
	}
	var req SignUpReq
	//Bind方法根据Content-Type来解析数据到req里面
	//解析错了直接返回一个400的错误
	if err := ctx.Bind(&req); err != nil {
		return
	}
	//校验邮箱
	ok, _ := u.emailExp.MatchString(req.Email)
	if !ok {
		ctx.String(http.StatusOK, "你的邮箱格式不对")
	}
	//校验密码
	if req.Password != req.ConfirmPassword {
		ctx.String(http.StatusOK, "两次输入的密码不一致")
		return
	}
	ok, _ = u.passwordExp.MatchString(req.Password)
	if !ok {
		ctx.String(http.StatusOK, "密码必须大于8位，包含数字、字母、特殊字符")
		return
	}
	//调用service层
	err := u.svc.SignUp(ctx, domain.User{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		if err == service.ErrUserDuplicateEmail {
			ctx.String(http.StatusOK, "邮箱冲突")
		} else {
			ctx.String(http.StatusOK, "系统异常")
		}
		return
	}
	ctx.String(http.StatusOK, "注册成功")
}
func (u *UserHandler) Login(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	user, err := u.svc.Login(ctx, req.Email, req.Password)
	if err != nil {
		if err == service.ErrInvalidUserOrPassword {
			ctx.String(http.StatusOK, "用户不存在或密码不对")
		} else {
			ctx.String(http.StatusOK, "系统异常")
		}
		return
	}
	//用jwt设置登录状态
	//创建token
	claims := middleware.UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			//设置过期时间
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
		},
		UserAgent: ctx.Request.UserAgent(),
		Uid:       user.Id,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte("secret"))
	if err != nil {
		ctx.String(http.StatusInternalServerError, "系统错误")
		return
	}
	//将token放到header中
	ctx.Header("x-jwt-token", tokenStr)
	//fmt.Println(tokenStr)
	ctx.String(http.StatusOK, "登录成功")
}

func (u *UserHandler) Login2(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	user, err := u.svc.Login(ctx, req.Email, req.Password)
	if err != nil {
		if err == service.ErrInvalidUserOrPassword {
			ctx.String(http.StatusOK, "用户不存在或密码不对")
		} else {
			ctx.String(http.StatusOK, "系统异常")
		}
		return
	}
	//登录成功
	//获取ctx中创建的session，设置session的键值对为userId:user.Id
	sess := sessions.Default(ctx)
	sess.Set("userId", user.Id)
	sess.Options(sessions.Options{
		//60秒过期
		MaxAge: 60,
	})
	sess.Save()
	ctx.String(http.StatusOK, "登录成功")
}

// Logout 登出
func (u *UserHandler) Logout(ctx *gin.Context) {
	sess := sessions.Default(ctx)
	sess.Options(sessions.Options{
		//删除用户的cookie
		MaxAge: -1,
	})
	sess.Save()
	ctx.String(http.StatusOK, "退出登录成功")
}

func (u *UserHandler) Edit(ctx *gin.Context) {
}

func (u *UserHandler) Profile(ctx *gin.Context) {
	c, ok := ctx.Get("claims")
	//一定获取到claims
	if !ok {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	//类型断言
	claims, ok := c.(*middleware.UserClaims)
	if !ok {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	fmt.Println(claims.Uid)
}

func (u *UserHandler) Profile2(ctx *gin.Context) {
}

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
	phoneExp    *regexp2.Regexp //手机号码校验器
	userSvc     service.UserService
	codeSvc     service.CodeService
}

func NewUserHandler(userSvc service.UserService, codeSvc service.CodeService) *UserHandler {
	//用于校验密码和邮箱的正则表达式
	const (
		emailRegexPattern = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
		// 和上面比起来，用 ` 看起来就比较清爽
		passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
		phoneRegexPattern    = `^1\d{10}$`
	)
	//初始化校验器
	return &UserHandler{
		userSvc:     userSvc,
		codeSvc:     codeSvc,
		emailExp:    regexp2.MustCompile(emailRegexPattern, regexp2.None),
		passwordExp: regexp2.MustCompile(passwordRegexPattern, regexp2.None),
		phoneExp:    regexp2.MustCompile(phoneRegexPattern, regexp2.None),
	}
}

func (u *UserHandler) RegisterRoutes(r *gin.Engine) {
	r.POST("users/signup", u.SignUp)
	r.POST("users/login", u.Login)
	r.POST("users/edit", u.Edit)
	r.POST("users/profile", u.Profile)
	//发送验证码
	r.POST("/users/login_sms/code/send", u.SendLoginSMSCode)
	//校验验证码
	r.POST("/login_sms", u.LoginSMS)
}

func (u *UserHandler) LoginSMS(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	//通过正则表达式判断是否是一个合法的电话号码
	ok, _ := u.phoneExp.MatchString(req.Phone)
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "输入有误",
		})
		return
	}
	//这里也可以先初步校验一下验证码
	ok, err := u.codeSvc.Verify(ctx, "login", req.Phone, req.Code)
	if err != nil {
		var msg string
		if err == service.ErrCodeVerifyTooMany {
			msg = "验证次数耗尽"
		} else {
			msg = "系统错误"
		}
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  msg,
		})
		return
	}
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "验证有误",
		})
		return
	}
	user, err := u.userSvc.FindOrCreate(ctx, req.Phone)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
	}
	//用jwt设置登录状态
	//创建并设置token
	if err := u.setJWTTloken(ctx, user.Id); err != nil {
		ctx.String(http.StatusInternalServerError, "系统错误")
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Code: 4,
		Msg:  "登录成功",
	})
}

func (u *UserHandler) SendLoginSMSCode(ctx *gin.Context) {
	//获取手机号码
	type Req struct {
		Phone string `json:"phone"`
	}
	var req Req
	//解析错了直接返回一个400的错误
	if err := ctx.Bind(&req); err != nil {
		return
	}
	//调用短信服务发送短信
	err := u.codeSvc.Send(ctx, "login", req.Phone)
	if err != nil {
		if err == service.ErrCodeSendTooMany {
			ctx.JSON(http.StatusOK, Result{
				Code: 4,
				Msg:  "短信发送太频繁，请稍后再试",
			})
		} else if err == service.ErrCodeVerifyTooMany {
			ctx.JSON(http.StatusOK, Result{
				Code: 4,
				Msg:  "验证太频繁",
			})
		} else {
			fmt.Println(err)
			ctx.JSON(http.StatusOK, Result{
				Code: 4,
				Msg:  "系统错误",
			})
		}
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "发送成功",
	})
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
		return
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
	err := u.userSvc.SignUp(ctx, domain.User{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		if err == service.ErrUserDuplicate {
			ctx.String(http.StatusOK, "邮箱冲突")
		} else {
			ctx.String(http.StatusOK, "系统异常")
		}
		return
	}
	ctx.String(http.StatusOK, "注册成功")
}

// Login 使用jwt登录
func (u *UserHandler) Login(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	user, err := u.userSvc.Login(ctx, req.Email, req.Password)
	if err != nil {
		if err == service.ErrInvalidUserOrPassword {
			ctx.String(http.StatusOK, "用户不存在或密码不对")
		} else {
			ctx.String(http.StatusOK, "系统异常")
		}
		return
	}
	//用jwt设置登录状态
	//创建并设置token
	if err := u.setJWTTloken(ctx, user.Id); err != nil {
		ctx.String(http.StatusInternalServerError, "系统错误")
		return
	}
	//fmt.Println(tokenStr)
	ctx.String(http.StatusOK, "登录成功")
}

func (u *UserHandler) setJWTTloken(ctx *gin.Context, uid int64) error {
	claims := middleware.UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			//设置过期时间
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
		},
		UserAgent: ctx.Request.UserAgent(),
		Uid:       uid,
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

func (u *UserHandler) Login2(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	user, err := u.userSvc.Login(ctx, req.Email, req.Password)
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
	// 嵌入一段刷新过期时间的代码
	type Req struct {
		// 改邮箱，密码，或者能不能改手机号
		Nickname string `json:"nickname"`
		// YYYY-MM-DD
		Birthday string `json:"birthday"`
		AboutMe  string `json:"aboutMe"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	//sess := sessions.Default(ctx)
	//sess.Get("uid")
	uc, ok := ctx.MustGet("claims").(middleware.UserClaims)
	if !ok {
		//ctx.String(http.StatusOK, "系统错误")
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	// 用户输入不对
	birthday, err := time.Parse(time.DateOnly, req.Birthday)
	if err != nil {
		//ctx.String(http.StatusOK, "系统错误")
		ctx.String(http.StatusOK, "生日格式不对")
		return
	}
	err = u.userSvc.UpdateNonSensitiveInfo(ctx, domain.User{
		Id:       uc.Uid,
		Nickname: req.Nickname,
		Birthday: birthday,
		AboutMe:  req.AboutMe,
	})
	if err != nil {
		ctx.String(http.StatusOK, "系统异常")
		return
	}
	ctx.String(http.StatusOK, "更新成功")
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
	user, err := u.userSvc.Profile(ctx, claims.Uid)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(user)
}

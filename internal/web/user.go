package web

import (
	"badminton-backend/internal/domain"
	"badminton-backend/internal/service"
	ijwt "badminton-backend/internal/web/jwt"
	"errors"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"net/http"
	"time"
)

const (
	phoneRegexPattern = `^1\d{10}$`

	passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`

	bizLogin = "login"
)

var _ handler = &UserHandler{}

// UserHandler 用于处理用户相关的HTTP请求
type UserHandler struct {
	svc              service.UserService
	codeSvc          service.CodeService
	phoneRegexExp    *regexp.Regexp
	passwordRegexExp *regexp.Regexp

	ijwt.Handler
}

func NewUserHandler(svc service.UserService, codeSvc service.CodeService, jwthdl ijwt.Handler) *UserHandler {
	return &UserHandler{
		svc:              svc,
		codeSvc:          codeSvc,
		phoneRegexExp:    regexp.MustCompile(phoneRegexPattern, regexp.None),
		passwordRegexExp: regexp.MustCompile(passwordRegexPattern, regexp.None),
		Handler:          jwthdl,
	}
}

func (c *UserHandler) RegisterRoutes(server *gin.Engine) {

	v1 := server.Group("/api/v1")
	ug := v1.Group("/user")

	ug.POST("/signup", c.SignUp)
	ug.POST("/login", c.Login)
	ug.POST("/logout", c.Logout)
	ug.POST("/edit", c.Edit)
	ug.GET("/profile", c.Profile)

	ug.POST("/login_sms/code/send", c.SendSMSLoginCode)
	ug.POST("/login_sms", c.LoginSMS)
	ug.POST("/refresh_token", c.RefreshToken)
}

func (c *UserHandler) RefreshToken(ctx *gin.Context) {
	tokenStr := c.ExtractTokenString(ctx)
	var rc ijwt.RefreshClaims
	token, err := jwt.ParseWithClaims(tokenStr, &rc, func(token *jwt.Token) (interface{}, error) {
		return ijwt.RefreshTokenKey, nil
	})
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, Result{
			Code: 14001,
			Msg:  "请登录",
		})
		return
	}
	if token == nil || !token.Valid {
		ctx.JSON(http.StatusUnauthorized, Result{
			Code: 14001,
			Msg:  "请登录",
		})
		return
	}

	err = c.CheckSession(ctx, rc.Ssid)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, Result{
			Code: 14001,
			Msg:  "请登录",
		})
		return
	}

	err = c.SetJWTToken(ctx, rc.Ssid, rc.Id)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, Result{
			Code: 14001,
			Msg:  "请登录",
		})
		return
	}
	ctx.JSON(http.StatusUnauthorized, Result{
		Code: 10000,
		Msg:  "OK",
	})
}

func (c *UserHandler) LoginSMS(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 25001,
			Msg:  "服务异常",
		})
		return
	}
	ok, err := c.codeSvc.Verify(ctx, bizLogin, req.Phone, req.Code)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 25001,
			Msg:  "服务异常",
		})
		return
	}
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 14002,
			Msg:  "验证码错误",
		})
		return
	}

	// 验证码是对的
	// 登录或者注册用户
	u, err := c.svc.FindOrCreate(ctx, req.Phone)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 25001,
			Msg:  "服务异常",
		})
		return
	}

	ssid := uuid.New().String()
	err = c.SetJWTToken(ctx, ssid, u.Id)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 25001,
			Msg:  "服务异常",
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Code: 10000,
		Msg:  "OK",
	})
}

func (c *UserHandler) SendSMSLoginCode(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 25001,
			Msg:  "服务异常",
		})
		return
	}
	if req.Phone == "" {
		ctx.JSON(http.StatusOK, Result{
			Code: 14002,
			Msg:  "手机号不能为空",
		})
		return
	}
	isPhone, err := c.phoneRegexExp.MatchString(req.Phone)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 25001,
			Msg:  "服务异常",
		})
		return
	}
	if !isPhone {
		ctx.JSON(http.StatusOK, Result{
			Code: 14002,
			Msg:  "手机号格式不正确",
		})
		return
	}

	err = c.codeSvc.Send(ctx, bizLogin, req.Phone)
	switch {
	case err == nil:
		ctx.JSON(http.StatusOK, Result{
			Code: 10000,
			Msg:  "OK",
		})

	case errors.Is(err, service.ErrCodeSendTooMany):
		ctx.JSON(http.StatusOK, Result{
			Code: 14003,
			Msg:  "短信发送太频繁，请稍后再试",
		})

	default:
		ctx.JSON(http.StatusOK, Result{
			Code: 25001,
			Msg:  "服务异常",
		})
		return
	}
}

func (c *UserHandler) SignUp(ctx *gin.Context) {
	type SignUpReq struct {
		Account         string `json:"account"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
	}
	var req SignUpReq
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 25001,
			Msg:  "服务异常",
		})
		return
	}

	if req.Password != req.ConfirmPassword {
		ctx.JSON(http.StatusOK, Result{
			Code: 14002,
			Msg:  "两次输入的密码不相同",
		})
		return
	}
	isPassword, err := c.passwordRegexExp.MatchString(req.Password)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 25001,
			Msg:  "服务异常",
		})
		return
	}
	if !isPassword {
		ctx.JSON(http.StatusOK, Result{
			Code: 14002,
			Msg:  "密码必须包含字母、数字、特殊字符，并且长度不能小于 8 位",
		})
		return
	}

	err = c.svc.Signup(
		ctx.Request.Context(),
		domain.User{Account: req.Account, Password: req.ConfirmPassword})
	if errors.Is(err, service.ErrUserDuplicateEmail) {
		ctx.JSON(http.StatusOK, Result{
			Code: 14002,
			Msg:  "用户名已注册",
		})
		return
	}
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 25001,
			Msg:  "服务异常",
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Code: 10000,
		Msg:  "注册成功",
	})
}

func (c *UserHandler) Login(ctx *gin.Context) {
	type LoginReq struct {
		Account  string `json:"account"`
		Password string `json:"password"`
	}
	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 25001,
			Msg:  "服务异常",
		})
		return
	}

	u, err := c.svc.Login(ctx.Request.Context(), req.Account, req.Password)
	if errors.Is(err, service.ErrInvalidUserOrPassword) {
		ctx.JSON(http.StatusOK, Result{
			Code: 14001,
			Msg:  "用户名或者密码不正确，请重试",
		})
		return
	}

	err = c.SetLoginToken(ctx, u.Id)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 25001,
			Msg:  "服务异常",
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Code: 10000,
		Msg:  "登录成功",
	})
}

func (c *UserHandler) Logout(ctx *gin.Context) {
	err := c.ClearToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 25001,
			Msg:  "服务异常",
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Code: 10000,
		Msg:  "退出登录成功",
	})
}

func (c *UserHandler) Edit(ctx *gin.Context) {
	type Req struct {
		Username string `json:"username"`
		WeightKg int    `json:"weightKg"`
		HeightCm int    `json:"heightCm"`
		Nickname string `json:"nickname"`
		Birthday string `json:"birthday"`
		AboutMe  string `json:"aboutMe"`
	}

	var req Req
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 25001,
			Msg:  "服务异常",
		})
		return
	}

	if len(req.AboutMe) > 1024 {
		ctx.JSON(http.StatusOK, Result{
			Code: 14002,
			Msg:  "自我介绍过长"})
		return
	}
	var birthday time.Time
	if req.Birthday != "" {
		var err error
		birthday, err = time.Parse(time.DateOnly, req.Birthday)
		if err != nil {
			ctx.JSON(http.StatusOK, Result{
				Code: 14002,
				Msg:  "日期格式不对",
			})
			return
		}
	}

	uc := ctx.MustGet("user").(ijwt.UserClaims)
	err := c.svc.UpdateNonSensitiveInfo(ctx, domain.User{
		Id:       uc.Id,
		Username: req.Username,
		Nickname: req.Nickname,
		WeightKG: req.WeightKg,
		HeightCM: req.HeightCm,
		AboutMe:  req.AboutMe,
		Birthday: birthday,
	})
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 25001,
			Msg:  "服务异常",
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Code: 1000,
		Msg:  "更新成功",
	})
	return
}

func (c *UserHandler) Profile(ctx *gin.Context) {

	type Profile struct {
		Username string
		Account  string
		Phone    string
		WeightKg int
		HeightCm int
		Nickname string
		Birthday string
		AboutMe  string
	}

	uc := ctx.MustGet("user").(ijwt.UserClaims)

	u, err := c.svc.Profile(ctx, uc.Id)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 25001,
			Msg:  "服务异常",
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Code: 1000,
		Msg:  "OK",
		Data: Profile{
			Username: u.Username,
			Account:  u.Account,
			Phone:    u.Phone,
			WeightKg: u.WeightKG,
			HeightCm: u.HeightCM,
			Nickname: u.Nickname,
			Birthday: u.Birthday.Format(time.DateOnly),
			AboutMe:  u.AboutMe,
		},
	})
}

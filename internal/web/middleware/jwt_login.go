package middleware

import (
	ijwt "badminton-backend/internal/web/jwt"
	"github.com/ecodeclub/ekit/set"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"time"
)

// JWTLoginMiddlewareBuilder 是一个中间件构建器，用于验证用户请求中的JWT令牌。
type JWTLoginMiddlewareBuilder struct {
	publicPaths set.Set[string]
	ijwt.Handler
}

func NewJWTLoginMiddlewareBuilder(hdl ijwt.Handler) *JWTLoginMiddlewareBuilder {
	s := set.NewMapSet[string](5)
	// 如果请求的路径是用户注册（/users/signup）或登录（/users/login）
	// 这些接口不需要JWT验证，直接放行
	s.Add("/api/v1/user/signup")
	s.Add("/api/v1/user/login_sms/code/send")
	s.Add("/api/v1/user/login_sms")
	s.Add("/api/v1/user/login")
	s.Add("/api/v1/user/refresh_token")
	return &JWTLoginMiddlewareBuilder{
		publicPaths: s,
		Handler:     hdl,
	}
}

// Build 方法创建并返回一个Gin的中间件，负责JWT的验证。
func (j *JWTLoginMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 不需要校验
		if j.publicPaths.Exist(ctx.Request.URL.Path) {
			return
		}

		tokenStr := j.ExtractTokenString(ctx)

		// 创建UserClaims结构体用于解析token中的claim信息
		uc := ijwt.UserClaims{}

		// 使用jwt.ParseWithClaims解析token并验证其合法性
		// tokenStr是待验证的JWT字符串，uc是用于存放解析后的claims信息
		// web.JWTKey是密钥，用于验证token的签名
		token, err := jwt.ParseWithClaims(tokenStr, &uc, func(token *jwt.Token) (interface{}, error) {
			return ijwt.AccessTokenKey, nil
		})

		if err != nil || !token.Valid {
			// 如果token解析失败或无效，返回401 Unauthorized
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 从UserClaims中获取过期时间（expiresAt）
		expireTime, err := uc.GetExpirationTime()
		if err != nil {
			// 如果无法获取过期时间，返回401 Unauthorized
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 如果token已经过期，返回401 Unauthorized
		if expireTime.Before(time.Now()) {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if ctx.GetHeader("User-Agent") != uc.UserAgent {
			// 换了一个 User-Agent，可能是攻击者
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		err = j.CheckSession(ctx, uc.Ssid)
		if err != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		ctx.Set("user", uc)
	}
}

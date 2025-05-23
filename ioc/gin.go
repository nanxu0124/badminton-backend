package ioc

import (
	"badminton-backend/internal/web"
	ijwt "badminton-backend/internal/web/jwt"
	"badminton-backend/internal/web/middleware"
	"badminton-backend/pkg/ginx/middleware/accesslog"
	"badminton-backend/pkg/ginx/middleware/ratelimit"
	"badminton-backend/pkg/logger"
	"context"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
)

func InitWebServer(funcs []gin.HandlerFunc, userHdl *web.UserHandler, summaryHdl *web.DailySummaryHandler) *gin.Engine {
	server := gin.Default() // 初始化一个默认的 Gin 引擎实例
	gin.ForceConsoleColor() // 强制开启控制台的彩色输出

	// 使用传入的中间件
	server.Use(funcs...)

	// 注册路由
	userHdl.RegisterRoutes(server)
	summaryHdl.RegisterRoutes(server)

	return server // 返回配置好的 Gin 引擎实例
}

func GinMiddlewares(cmd redis.Cmdable, hdl ijwt.Handler, l logger.Logger) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		ratelimit.NewBuilder(cmd, time.Minute, 100).Build(), // 限制每分钟最多 100 次请求
		corsHandler(), // 配置 CORS 中间件

		// 使用 JWT 中间件
		middleware.NewJWTLoginMiddlewareBuilder(hdl).Build(),

		// 访问日志中间件
		accesslog.NewMiddlewareBuilder(func(ctx context.Context, al accesslog.AccessLog) {
			// 设置为 DEBUG 级别
			l.Debug("GIN 收到请求", logger.Field{
				Key:   "req",
				Value: al,
			})
		}).AllowReqBody().AllowRespBody().Build(),
	}
}

func corsHandler() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowCredentials: true,                                       // 允许客户端发送认证信息
		AllowHeaders:     []string{"Content-Type", "Authorization"},  // 允许的请求头
		ExposeHeaders:    []string{"X-Jwt-Token", "X-Refresh-Token"}, // 暴露的响应头
		AllowOriginFunc: func(origin string) bool {
			// 允许来自 localhost 和指定公司域名的请求
			if strings.HasPrefix(origin, "http://localhost") {
				return true
			}
			return strings.Contains(origin, "baidu.com")
		},
		MaxAge: 12 * time.Hour, // 预检请求的缓存时间，12小时内不会再进行预检
	})
}

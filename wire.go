//go:build wireinject

package main

import (
	"badminton-backend/internal/repository"
	"badminton-backend/internal/repository/cache"
	"badminton-backend/internal/repository/dao"
	"badminton-backend/internal/service"
	"badminton-backend/internal/web"
	ijwt "badminton-backend/internal/web/jwt"
	"badminton-backend/ioc"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

func InitWebServer() *gin.Engine {
	wire.Build(
		ioc.InitDB, ioc.InitRedis,

		dao.NewGormUserDAO,
		dao.NewGormDailySummaryDAO,

		cache.NewRedisUserCache,
		cache.NewRedisCodeCache,
		cache.NewRedisDailySummaryCache,

		repository.NewCachedUserRepository,
		repository.NewCachedCodeRepository,
		repository.NewDailySummaryRepository,

		service.NewUserService,
		service.NewSMSCodeService,
		service.NewDailySummaryService,

		ioc.GinMiddlewares,
		ioc.InitWebServer,
		ioc.InitLogger,
		ioc.InitSmsService,
		ijwt.NewRedisHandler,

		web.NewUserHandler,
		web.NewDailySummaryHandler,
	)

	return new(gin.Engine)
}

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

		cache.NewRedisUserCache, cache.NewRedisCodeCache,

		repository.NewCachedUserRepository,
		repository.NewCachedCodeRepository,

		service.NewUserService,
		service.NewSMSCodeService,
		ioc.InitSmsService,

		ijwt.NewRedisHandler,
		web.NewUserHandler,
		ioc.GinMiddlewares,
		ioc.InitWebServer,
		ioc.InitLogger,
	)

	return new(gin.Engine)
}

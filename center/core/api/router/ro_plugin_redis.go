package router

import (
	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/core/api/entry"
	"github.com/sensdata/idb/center/core/api/middleware"
)

type RedisRouter struct{}

func (s *RedisRouter) InitRouter(Router *gin.RouterGroup) {
	redisRouter := Router.Group("redis")
	redisRouter.Use(middleware.NewJWT().JWTAuth())
	baseApi := entry.ApiGroup
	{
		redisRouter.GET("/:host", baseApi.RedisComposes)
		redisRouter.POST("/:host/operation", baseApi.RedisOperation)
		redisRouter.POST("/:host/port", baseApi.RedisSetPort)
		redisRouter.GET("/:host/conf", baseApi.RedisGetConf)
		redisRouter.POST("/:host/conf", baseApi.RedisSetConf)
	}
}

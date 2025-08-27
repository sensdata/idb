package router

import (
	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/core/api/entry"
	"github.com/sensdata/idb/center/core/api/middleware"
)

type PostgreSqlRouter struct{}

func (s *PostgreSqlRouter) InitRouter(Router *gin.RouterGroup) {
	postgresqlRouter := Router.Group("postgresql")
	postgresqlRouter.Use(middleware.NewJWT().JWTAuth())
	baseApi := entry.ApiGroup
	{
		postgresqlRouter.GET("/:host", baseApi.PostgreSqlComposes)
		postgresqlRouter.POST("/:host/operation", baseApi.PostgreSqlOperation)
		postgresqlRouter.POST("/:host/port", baseApi.PostgreSqlSetPort)
		postgresqlRouter.GET("/:host/conf", baseApi.PostgreSqlGetConf)
		postgresqlRouter.POST("/:host/conf", baseApi.PostgreSqlSetConf)
	}
}

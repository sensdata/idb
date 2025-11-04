package router

import (
	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/core/api/entry"
	"github.com/sensdata/idb/center/core/api/middleware"
)

type PmaRouter struct{}

func (s *PmaRouter) InitRouter(Router *gin.RouterGroup) {
	pmaRouter := Router.Group("pma")
	pmaRouter.Use(middleware.NewJWT().JWTAuth())
	baseApi := entry.ApiGroup
	{
		pmaRouter.GET("/:host", baseApi.PmaComposes)
		pmaRouter.POST("/:host/operation", baseApi.PmaOperation)
		pmaRouter.POST("/:host/port", baseApi.PmaSetPort)
		pmaRouter.GET("/:host/servers", baseApi.PmaGetServers)
		pmaRouter.POST("/:host/server", baseApi.PmaAddServer)
		pmaRouter.PUT("/:host/server", baseApi.PmaUpdateServer)
		pmaRouter.DELETE("/:host/server", baseApi.PmaRemoveServer)
	}
}

package router

import (
	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/core/api/entry"
	"github.com/sensdata/idb/center/core/api/middleware"
)

type HomeRouter struct{}

func (s *HomeRouter) InitRouter(Router *gin.RouterGroup) {
	appRouter := Router.Group("home")
	appRouter.Use(middleware.NewJWT().JWTAuth())
	baseApi := entry.ApiGroup
	{
		appRouter.GET("/:host/managed/apps", baseApi.ManagedApps) //获取管理应用列表
	}
}

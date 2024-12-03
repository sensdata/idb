package router

import (
	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/core/api/entry"
	"github.com/sensdata/idb/center/core/api/middleware"
)

type AppRouter struct{}

func (s *AppRouter) InitRouter(Router *gin.RouterGroup) {
	appRouter := Router.Group("store")
	appRouter.Use(middleware.NewJWT().JWTAuth())
	baseApi := entry.ApiGroup
	{
		appRouter.POST("/apps/sync", baseApi.SyncApp)                    // 同步Apps
		appRouter.GET("/apps", baseApi.AppPage)                          // 获取应用列表
		appRouter.GET("/apps/:id", baseApi.AppDetail)                    // 获取应用详情
		appRouter.GET("/:host/apps/installed", baseApi.InstalledAppPage) //获取已安装应用列表
		appRouter.POST("/:host/apps/install", baseApi.InstallApp)        // 安装应用
	}
}

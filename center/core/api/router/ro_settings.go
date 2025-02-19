package router

import (
	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/core/api/entry"
	"github.com/sensdata/idb/center/core/api/middleware"
)

type SettingsRouter struct{}

func (s *SettingsRouter) InitRouter(Router *gin.RouterGroup) {
	settingsRouter := Router.Group("settings")
	settingsRouter.Use(middleware.NewJWT().JWTAuth())
	baseApi := entry.ApiGroup
	{
		settingsRouter.GET("/about", baseApi.About)
		settingsRouter.GET("/ips", baseApi.IPs)
		settingsRouter.GET("", baseApi.Settings)
		settingsRouter.POST("", baseApi.UpdateSettings)
	}
}

package router

import (
	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/core/api/entry"
	"github.com/sensdata/idb/center/core/api/middleware"
)

type HostRouter struct{}

func (s *HostRouter) InitRouter(Router *gin.RouterGroup) {
	hostRouter := Router.Group("hosts")
	hostRouter.Use(middleware.NewJWT().JWTAuth())
	baseApi := entry.ApiGroup
	{
		hostRouter.GET("/groups", baseApi.ListHostGroup)              // 获取设备组列表
		hostRouter.GET("", baseApi.ListHost)                          // 获取设备列表
		hostRouter.POST("", baseApi.CreateHost)                       // 新增设备
		hostRouter.PUT("", baseApi.UpdateHost)                        // 更新设备
		hostRouter.PUT("/:host/ssh", baseApi.UpdateHostSSH)           // 更新设备ssh配置
		hostRouter.PUT("/:host/agent", baseApi.UpdateHostAgent)       // 更新设备agent配置
		hostRouter.POST("/:host/test/ssh", baseApi.TestHostSSH)       // 测试设备ssh
		hostRouter.POST("/:host/test/agent", baseApi.TestHostAgent)   // 测试设备agent
		hostRouter.POST("/:host/install/agent", baseApi.InstallAgent) // 安装agent
	}
}

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
		hostRouter.GET("/groups", baseApi.ListHostGroup)                        // 获取设备组列表
		hostRouter.POST("/groups", baseApi.CreateHostGroup)                     // 创建设备组
		hostRouter.PUT("/groups", baseApi.UpdateHostGroup)                      // 更新设备组
		hostRouter.DELETE("/groups", baseApi.DeleteHostGroup)                   // 删除设备组
		hostRouter.GET("", baseApi.ListHost)                                    // 获取设备列表
		hostRouter.POST("", baseApi.CreateHost)                                 // 新增设备
		hostRouter.PUT("", baseApi.UpdateHost)                                  // 更新设备
		hostRouter.DELETE("", baseApi.DeleteHost)                               // 删除设备
		hostRouter.GET("/:host/status", baseApi.HostStatus)                     // 设备状态
		hostRouter.GET("/:host/status/follow", baseApi.HostStatusFollow)        // 追踪设备状态
		hostRouter.POST("/:host/activate", baseApi.ActivateHost)                // 激活设备
		hostRouter.GET("/:host", baseApi.HostInfo)                              // 设备配置信息
		hostRouter.PUT("/:host/conf/ssh", baseApi.UpdateHostSSH)                // 更新设备ssh配置
		hostRouter.PUT("/:host/conf/agent", baseApi.UpdateHostAgent)            // 更新设备agent配置
		hostRouter.POST("/test/ssh", baseApi.TestHostSSH)                       // 测试设备ssh
		hostRouter.POST("/:host/test/agent", baseApi.TestHostAgent)             // 测试设备agent
		hostRouter.POST("/:host/agent/install", baseApi.InstallAgent)           // 安装agent
		hostRouter.POST("/:host/agent/uninstall", baseApi.UninstallAgent)       // 卸载agent
		hostRouter.GET("/:host/agent/status", baseApi.AgentStatus)              // 获取agent状态
		hostRouter.GET("/:host/agent/status/follow", baseApi.AgentStatusFollow) // 追踪agent状态
		hostRouter.POST("/:host/agent/restart", baseApi.RestartAgent)           // 重启agent
	}
}

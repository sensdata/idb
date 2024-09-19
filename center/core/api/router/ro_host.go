package router

import (
	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/core/api/entry"
	"github.com/sensdata/idb/center/core/api/middleware"
)

type HostRouter struct{}

func (s *HostRouter) InitRouter(Router *gin.RouterGroup) {
	hostRouter := Router.Group("host")
	hostRouter.Use(middleware.NewJWT().JWTAuth())
	baseApi := entry.ApiGroup
	{
		hostRouter.POST("/group/list", baseApi.ListHostGroup)
		hostRouter.POST("/list", baseApi.ListHost)
		hostRouter.POST("/create", baseApi.CreateHost)
		hostRouter.POST("/update", baseApi.UpdateHost)
		hostRouter.POST("/update/ssh", baseApi.UpdateHostSSH)
		hostRouter.POST("/update/agent", baseApi.UpdateHostAgent)
		hostRouter.POST("/test/ssh", baseApi.TestHostSSH)
		hostRouter.POST("/test/agent", baseApi.TestHostAgent)
	}
}

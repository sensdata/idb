package router

import (
	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/core/api/entry"
	"github.com/sensdata/idb/center/core/api/middleware"
)

type RsyncRouter struct{}

func (s *RsyncRouter) InitRouter(Router *gin.RouterGroup) {
	rsyncRouter := Router.Group("rsync")
	rsyncRouter.Use(middleware.NewJWT().JWTAuth())
	baseApi := entry.ApiGroup
	{
		rsyncRouter.GET("/task", baseApi.RsyncListTask)
		rsyncRouter.GET("/task/query", baseApi.RsyncQueryTask)
		rsyncRouter.POST("/task", baseApi.RsyncCreateTask)
		rsyncRouter.DELETE("/task", baseApi.RsyncDeleteTask)
		rsyncRouter.POST("/task/cancel", baseApi.RsyncCancelTask)
		rsyncRouter.POST("/task/retry", baseApi.RsyncRetryTask)
	}
}

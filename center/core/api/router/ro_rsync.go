package router

import (
	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/core/api/entry"
	"github.com/sensdata/idb/center/core/api/middleware"
)

type RsyncClientRouter struct{}

func (s *RsyncClientRouter) InitRouter(Router *gin.RouterGroup) {
	rsyncRouter := Router.Group("rsync")
	rsyncRouter.Use(middleware.NewJWT().JWTAuth())
	baseApi := entry.ApiGroup
	{
		rsyncRouter.GET("/:host/task", baseApi.RsyncListTask)
		rsyncRouter.GET("/:host/task/query", baseApi.RsyncQueryTask)
		rsyncRouter.POST("/:host/task", baseApi.RsyncCreateTask)
		rsyncRouter.DELETE("/:host/task", baseApi.RsyncDeleteTask)
		rsyncRouter.POST("/:host/task/cancel", baseApi.RsyncCancelTask)
		rsyncRouter.POST("/:host/task/retry", baseApi.RsyncRetryTask)
	}
}

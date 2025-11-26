package router

import (
	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/core/api/entry"
	"github.com/sensdata/idb/center/core/api/middleware"
)

type RsyncRouter struct{}

func (s *RsyncRouter) InitRouter(Router *gin.RouterGroup) {
	rsyncRouter := Router.Group("transfer")
	rsyncRouter.Use(middleware.NewJWT().JWTAuth())
	baseApi := entry.ApiGroup
	{
		rsyncRouter.GET("/task", baseApi.TransferListTask)
		rsyncRouter.GET("/task/query", baseApi.TransferQueryTask)
		rsyncRouter.POST("/task", baseApi.TransferCreateTask)
		rsyncRouter.DELETE("/task", baseApi.TransferDeleteTask)
		rsyncRouter.POST("/task/cancel", baseApi.TransferCancelTask)
		rsyncRouter.POST("/task/retry", baseApi.TransferRetryTask)
	}
}

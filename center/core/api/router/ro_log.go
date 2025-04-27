package router

import (
	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/core/api/entry"
	"github.com/sensdata/idb/center/core/api/middleware"
)

type LogManRouter struct{}

func (s *LogManRouter) InitRouter(Router *gin.RouterGroup) {
	taskRouter := Router.Group("logs")
	baseApi := entry.ApiGroup
	{
		taskRouter.GET("/:host/follow", middleware.NewJWT().JWTAuth(), baseApi.HandleLogStream) // 连接到日志流
	}
}

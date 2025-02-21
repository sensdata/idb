package router

import (
	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/core/api/entry"
	"github.com/sensdata/idb/center/core/api/middleware"
)

type TaskRouter struct{}

func (s *TaskRouter) InitRouter(Router *gin.RouterGroup) {
	taskRouter := Router.Group("tasks")
	baseApi := entry.ApiGroup
	{
		taskRouter.GET("/:taskId/logs/stream", middleware.NewJWT().JWTCookieAuth(), baseApi.HandleTaskLogStream) // 连接到任务日志流
	}
}

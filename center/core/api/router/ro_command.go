package router

import (
	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/core/api/entry"
)

type CommandRouter struct{}

func (s *CommandRouter) InitRouter(Router *gin.RouterGroup) {
	commandRouter := Router.Group("commands")
	baseApi := entry.ApiGroup
	{
		commandRouter.POST("", baseApi.SendCommand)            // 发送单个命令
		commandRouter.POST("/group", baseApi.SendCommandGroup) // 发送一组命令
	}
}

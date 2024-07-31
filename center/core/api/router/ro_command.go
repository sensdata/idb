package router

import (
	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/core/api/entry"
)

type CommandRouter struct{}

func (s *CommandRouter) InitRouter(Router *gin.RouterGroup) {
	commandRouter := Router.Group("cmd")
	baseApi := entry.ApiGroup
	{
		commandRouter.POST("/send", baseApi.SendCommand)
	}
}

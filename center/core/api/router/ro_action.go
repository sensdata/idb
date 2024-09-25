package router

import (
	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/core/api/entry"
)

type ActionRouter struct{}

func (s *ActionRouter) InitRouter(Router *gin.RouterGroup) {
	actionRouter := Router.Group("actions")
	baseApi := entry.ApiGroup
	{
		actionRouter.POST("", baseApi.SendAction) // 向目标设备发送action指令
	}
}

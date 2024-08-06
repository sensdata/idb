package router

import (
	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/core/api/entry"
)

type ActionRouter struct{}

func (s *ActionRouter) InitRouter(Router *gin.RouterGroup) {
	actionRouter := Router.Group("act")
	baseApi := entry.ApiGroup
	{
		actionRouter.POST("/send", baseApi.SendAction)
	}
}

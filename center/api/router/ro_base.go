package router

import (
	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/api/entry"
)

type BaseRouter struct{}

func (s *BaseRouter) InitRouter(Router *gin.RouterGroup) {
	baseRouter := Router.Group("auth")
	baseApi := entry.ApiGroup
	{
		baseRouter.POST("/login", baseApi.Login)
		baseRouter.POST("/logout", baseApi.Logout)
	}
}

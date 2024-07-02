package router

import (
	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/core/api/entry"
)

type AuthRouter struct{}

func (s *AuthRouter) InitRouter(Router *gin.RouterGroup) {
	baseRouter := Router.Group("auth")
	baseApi := entry.ApiGroup
	{
		baseRouter.POST("/login", baseApi.Login)
		baseRouter.GET("/logout", baseApi.Logout)
	}
}

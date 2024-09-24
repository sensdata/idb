package router

import (
	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/core/api/entry"
	"github.com/sensdata/idb/center/core/api/middleware"
)

type AuthRouter struct{}

func (s *AuthRouter) InitRouter(Router *gin.RouterGroup) {
	baseRouter := Router.Group("auth")
	baseApi := entry.ApiGroup
	{
		baseRouter.POST("/sessions", baseApi.Login)
		baseRouter.DELETE("/sessions", middleware.NewJWT().JWTAuth(), baseApi.Logout)
	}
}

package router

import (
	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/core/api/entry"
	"github.com/sensdata/idb/center/core/api/middleware"
)

type ScriptsRouter struct{}

func (s *ScriptsRouter) InitRouter(Router *gin.RouterGroup) {
	scriptRouter := Router.Group("pscripts")
	scriptRouter.Use(middleware.NewJWT().JWTAuth())
	baseApi := entry.ApiGroup
	{
		scriptRouter.GET("/:host", baseApi.GetScriptList)
	}
}

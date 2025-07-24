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
		scriptRouter.GET("/:host/category", baseApi.GetScriptCategories)
		scriptRouter.POST("/:host/category", baseApi.CreateScriptCategory)
		scriptRouter.PUT("/:host/category", baseApi.UpdateScriptCategory)
		scriptRouter.DELETE("/:host/category", baseApi.DeleteScriptCategory)
		scriptRouter.GET("/:host", baseApi.GetScriptList)
		scriptRouter.GET("/:host/detail", baseApi.GetScriptDetail)
		scriptRouter.POST("/:host", baseApi.CreateScript)
		scriptRouter.PUT("/:host", baseApi.UpdateScript)
		scriptRouter.DELETE("/:host", baseApi.DeleteScript)
		scriptRouter.PUT("/:host/restore", baseApi.RestoreScript)
		scriptRouter.GET("/:host/history", baseApi.GetScriptHistories)
		scriptRouter.GET("/:host/diff", baseApi.GetScriptDiff)
		scriptRouter.POST("/:host/sync", baseApi.ScriptSync)
		scriptRouter.POST("/:host/run", baseApi.ScriptExec)
		scriptRouter.GET("/:host/run/logs", baseApi.GetScriptRunLogs)
		scriptRouter.GET("/:host/run/logs/detail", baseApi.GetScriptRunLogDetail)
	}
}

package router

import (
	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/core/api/entry"
	"github.com/sensdata/idb/center/core/api/middleware"
)

type TerminalRouter struct{}

func (s *TerminalRouter) InitRouter(Router *gin.RouterGroup) {
	terminalRouter := Router.Group("ws")
	terminalRouter.Use(middleware.NewJWT().JWTAuth())
	baseApi := entry.ApiGroup
	{
		terminalRouter.GET("/terminals", baseApi.HandleTerminal) // WebSocket终端会话
	}
}

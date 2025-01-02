package router

import (
	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/core/api/entry"
)

type TerminalRouter struct{}

func (s *TerminalRouter) InitRouter(Router *gin.RouterGroup) {
	terminalRouter := Router.Group("ws")
	// terminalRouter.Use(middleware.NewJWT().JWTAuth())
	baseApi := entry.ApiGroup
	{
		terminalRouter.GET("/terminals/ssh", baseApi.HandleSshTerminal) // WebSocket终端会话
	}
}

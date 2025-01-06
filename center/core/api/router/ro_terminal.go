package router

import (
	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/core/api/entry"
	"github.com/sensdata/idb/center/core/api/middleware"
)

type TerminalRouter struct{}

func (s *TerminalRouter) InitRouter(Router *gin.RouterGroup) {
	terminalRouter := Router.Group("terminals")
	terminalRouter.Use(middleware.NewJWT().JWTCookieAuth())
	baseApi := entry.ApiGroup
	{
		// websocket接口
		terminalRouter.GET("/:host/ssh/start", baseApi.HandleSshTerminal) // 创建ssh终端会话

		terminalRouter.GET("/:host/start", baseApi.HandleTerminal) // 创建或连接到终端会话

		// http接口
		terminalRouter.GET("/:host/sessions", baseApi.TerminalSessions)     // 枚举终端会话
		terminalRouter.POST("/:host/session/detach", baseApi.DetachSession) // 分离终端会话
		terminalRouter.POST("/:host/session/quit", baseApi.QuitSession)     //终止终端会话
		terminalRouter.POST("/:host/session/rename", baseApi.RenameSession) // 重命名终端会话
		terminalRouter.POST("/:host/install", baseApi.InstallTerminal)      // 安装Agent侧的终端环境
	}
}

package router

import (
	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/core/api/entry"
	"github.com/sensdata/idb/center/core/api/middleware"
)

type TerminalRouter struct{}

func (s *TerminalRouter) InitRouter(Router *gin.RouterGroup) {
	terminalRouter := Router.Group("terminals")
	baseApi := entry.ApiGroup
	{
		// websocket接口
		terminalRouter.GET("/:host/ssh/start", middleware.NewJWT().JWTCookieAuth(), baseApi.HandleSshTerminal) // 创建ssh终端会话

		terminalRouter.GET("/:host/start", middleware.NewJWT().JWTCookieAuth(), baseApi.HandleTerminal) // 创建或连接到终端会话

		// http接口
		terminalRouter.GET("/:host/sessions", middleware.NewJWT().JWTAuth(), baseApi.TerminalSessions)     // 枚举终端会话
		terminalRouter.POST("/:host/session/detach", middleware.NewJWT().JWTAuth(), baseApi.DetachSession) // 分离终端会话
		terminalRouter.POST("/:host/session/quit", middleware.NewJWT().JWTAuth(), baseApi.QuitSession)     //终止终端会话
		terminalRouter.POST("/:host/session/rename", middleware.NewJWT().JWTAuth(), baseApi.RenameSession) // 重命名终端会话
		terminalRouter.POST("/:host/install", middleware.NewJWT().JWTAuth(), baseApi.InstallTerminal)      // 安装Agent侧的终端环境
	}
}

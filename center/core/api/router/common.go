package router

import "github.com/gin-gonic/gin"

var RouterGroups = commonGroups()

type CommonRouter interface {
	InitRouter(Router *gin.RouterGroup)
}

func commonGroups() []CommonRouter {
	return []CommonRouter{
		&AuthRouter{},
		&UserRouter{},
		&GroupRouter{},
		&HostRouter{},
		&CommandRouter{},
		&ActionRouter{},
		&TerminalRouter{},
		&AppRouter{},
		&SettingsRouter{},
		&LogManRouter{},
		&PublicRouter{},
		&ScriptsRouter{},
		&MysqlRouter{},
		&PostgreSqlRouter{},
		&RedisRouter{},
		&RsyncRouter{},
		&PmaRouter{},
	}
}

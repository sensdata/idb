package router

import "github.com/gin-gonic/gin"

var RouterGroups = commonGroups()
var WsGroups = wsGroups()

type CommonRouter interface {
	InitRouter(Router *gin.RouterGroup)
}

func commonGroups() []CommonRouter {
	return []CommonRouter{
		&AuthRouter{},
		&UserRouter{},
		&CommandRouter{},
		&ActionRouter{},
	}
}

func wsGroups() []CommonRouter {
	return []CommonRouter{
		&TerminalRouter{},
	}
}

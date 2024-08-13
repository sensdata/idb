package api

import (
	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/core/api/router"
	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/plugin"
	"github.com/sensdata/idb/core/utils"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var API ApiServer = ApiServer{
	router: gin.Default(),
}

type ApiServer struct {
	router *gin.Engine
}

func (s *ApiServer) InitRouter() {
	// 注册 API 路由
	s.setUpDefaultRouters()
	// 注册 WS 路由
	s.setUpWebSocketRouter()
}

func (s *ApiServer) Start() error {
	//初始化validator
	global.LOG.Info("Init validator")
	global.VALID = utils.InitValidator()

	err := s.router.Run("0.0.0.0:8080")
	if err != nil {
		global.LOG.Error("Failed to start HTTP server: %v\n", err)
	}

	return nil
}

// SetupRouter sets up the API routes
func (s *ApiServer) setUpDefaultRouters() {
	global.LOG.Info("register router - swagger")
	swaggerGroup := s.router.Group("idb/swagger")
	swaggerGroup.GET("/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	global.LOG.Info("register router - api")
	apiGroup := s.router.Group("idb/api")
	for _, router := range router.RouterGroups {
		router.InitRouter(apiGroup)
	}
}

// SetUpWebSocketRouter sets up web socket router
func (s *ApiServer) setUpWebSocketRouter() {
	global.LOG.Info("register router - websocket")
	wsGroup := s.router.Group("idb/ws")
	for _, router := range router.WsGroups {
		router.InitRouter(wsGroup)
	}
}

// SetUpPluginRouters sets up routers from plugins
func (s *ApiServer) SetUpPluginRouters(group string, routes []plugin.PluginRoute) {
	global.LOG.Info("register router - " + group)
	pluginGroup := s.router.Group("idb/" + group)
	for _, route := range routes {
		switch route.Method {
		case "GET":
			pluginGroup.GET(route.Path, route.Handler)
		case "POST":
			pluginGroup.POST(route.Path, route.Handler)
		}
	}
}

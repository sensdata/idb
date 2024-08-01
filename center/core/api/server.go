package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/core/api/router"
	"github.com/sensdata/idb/core/plugin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var API ApiServer = ApiServer{
	router: gin.Default(),
}

type ApiServer struct {
	router *gin.Engine
}

func (s *ApiServer) Start() error {
	// 注册 API 路由
	s.setUpDefaultRouters()

	// 仅监听本地接口
	err := s.router.Run("0.0.0.0:8080")
	if err != nil {
		fmt.Printf("Failed to start HTTP server: %v\n", err)
	}

	return nil
}

// SetupRouter sets up the API routes
func (s *ApiServer) setUpDefaultRouters() {
	swaggerGroup := s.router.Group("idb/swagger")
	swaggerGroup.GET("/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	apiGroup := s.router.Group("idb/api")
	for _, router := range router.RouterGroups {
		router.InitRouter(apiGroup)
	}
}

// SetUpPluginRouters sets up routers from plugins
func (s *ApiServer) SetUpPluginRouters(group string, routes []plugin.PluginRoute) {
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

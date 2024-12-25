package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/core/api/middleware"
	"github.com/sensdata/idb/center/core/api/router"
	"github.com/sensdata/idb/center/core/conn"
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
}

func (s *ApiServer) Start() error {
	//初始化validator
	global.LOG.Info("Init validator")
	global.VALID = utils.InitValidator()

	addr := fmt.Sprintf("%s:%d", "0.0.0.0", conn.CONFMAN.GetConfig().Port)
	err := s.router.Run(addr)
	if err != nil {
		global.LOG.Error("Failed to start HTTP server: %v\n", err)
	}

	return nil
}

// SetupRouter sets up the API routes
func (s *ApiServer) setUpDefaultRouters() {
	global.LOG.Info("register router - api")
	apiGroup := s.router.Group("api/v1")

	for _, router := range router.RouterGroups {
		router.InitRouter(apiGroup)
	}

	// 添加 Swagger 路由到 apiGroup
	global.LOG.Info("register router - swagger")
	swaggerGroup := apiGroup.Group("swagger")
	swaggerGroup.GET("/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	// 处理未匹配的请求，重定向到 /var/lib/idb/home
	s.router.NoRoute(func(c *gin.Context) {
		// 这里可以使用 c.FileServer 来处理目录下的所有请求
		c.File("/var/lib/idb/home/index.html") // 默认返回 index.html
	})

	// 设置静态文件路由，确保可以访问 assets 目录
	s.router.Static("/assets", "/var/lib/idb/home/assets") // 处理 assets 目录
}

// SetUpPluginRouters sets up routers from plugins
func (s *ApiServer) SetUpPluginRouters(group string, routes []plugin.PluginRoute) {
	global.LOG.Info("register router - " + group)
	pluginGroup := s.router.Group("api/v1/" + group)
	pluginGroup.Use(middleware.NewJWT().JWTAuth())
	for _, route := range routes {
		switch route.Method {
		case "GET":
			pluginGroup.GET(route.Path, route.Handler)
		case "POST":
			pluginGroup.POST(route.Path, route.Handler)
		case "DELETE":
			pluginGroup.DELETE(route.Path, route.Handler)
		case "PUT":
			pluginGroup.PUT(route.Path, route.Handler)
		}
	}
}

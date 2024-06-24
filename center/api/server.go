package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/api/router"
	"github.com/sensdata/idb/center/config"
	"github.com/sensdata/idb/center/core"
)

type ApiServer struct {
	cfg    config.CenterConfig
	center *core.Center
	router *gin.Engine
}

func NewApiServer(cfg config.CenterConfig, center *core.Center) *ApiServer {
	return &ApiServer{
		cfg:    cfg,
		center: center,
	}
}

func (s *ApiServer) Start() error {
	// 注册 API 端点和处理函数
	s.router = setupRouter()

	// 仅监听本地接口
	err := s.router.Run("127.0.0.1:8080")
	if err != nil {
		fmt.Printf("Failed to start HTTP server: %v\n", err)
	}

	return nil
}

// SetupRouter sets up the API routes
func setupRouter() *gin.Engine {
	Router := gin.Default()

	RouterGroup := Router.Group("/")
	for _, router := range router.RouterGroups {
		router.InitRouter(RouterGroup)
	}

	return Router
}

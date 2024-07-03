package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/core/api/router"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type ApiServer struct {
	router *gin.Engine
}

func NewApiServer() *ApiServer {
	return &ApiServer{}
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

	swaggerGroup := Router.Group("swagger")
	swaggerGroup.GET("/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	RouterGroup := Router.Group("/")
	for _, router := range router.RouterGroups {
		router.InitRouter(RouterGroup)
	}

	return Router
}

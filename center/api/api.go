package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/config"
	"github.com/sensdata/idb/center/core"
)

type Api struct {
	cfg    config.CenterConfig
	center *core.Center
	router *gin.Engine
}

func NewApi(cfg config.CenterConfig, center *core.Center) *Api {
	return &Api{
		cfg:    cfg,
		center: center,
		router: gin.Default(),
	}
}

func (s *Api) Start() error {
	// 注册 API 端点和处理函数
	s.router.GET("/test", s.testHandler)

	// 仅监听本地接口
	err := s.router.Run("127.0.0.1:8080")
	if err != nil {
		fmt.Printf("Failed to start HTTP server: %v\n", err)
	}

	return nil
}

func (s *Api) testHandler(c *gin.Context) {
	// 调用 Center 的 ExecuteCommand 方法
	result, err := s.center.ExecuteCommand("ps")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to execute command", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": result})
}

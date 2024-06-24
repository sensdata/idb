package api

import (
	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/api/router"
)

// SetupRouter sets up the API routes
func SetupRouter() *gin.Engine {
	Router := gin.Default()

	RouterGroup := Router.Group("/")
	for _, router := range router.RouterGroups {
		router.InitRouter(RouterGroup)
	}

	return Router
}

package router

import (
	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/core/api/entry"
)

type PublicRouter struct{}

func (s *PublicRouter) InitRouter(Router *gin.RouterGroup) {
	commonRouter := Router.Group("public")
	baseApi := entry.ApiGroup
	{
		commonRouter.GET("/version", baseApi.Version)
	}
}

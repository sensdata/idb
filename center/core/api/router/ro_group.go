package router

import (
	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/core/api/entry"
	"github.com/sensdata/idb/center/core/api/middleware"
)

type GroupRouter struct{}

func (s *GroupRouter) InitRouter(Router *gin.RouterGroup) {
	groupRouter := Router.Group("group")
	groupRouter.Use(middleware.NewJWT().JWTAuth())
	baseApi := entry.ApiGroup
	{
		groupRouter.POST("/list", baseApi.ListGroup)
		groupRouter.POST("/create", baseApi.CreateGroup)
		groupRouter.POST("/update", baseApi.UpdateGroup)
		groupRouter.POST("/delete", baseApi.DeleteGroup)
	}
}

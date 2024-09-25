package router

import (
	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/core/api/entry"
	"github.com/sensdata/idb/center/core/api/middleware"
)

type GroupRouter struct{}

func (s *GroupRouter) InitRouter(Router *gin.RouterGroup) {
	groupRouter := Router.Group("groups")
	groupRouter.Use(middleware.NewJWT().JWTAuth())
	baseApi := entry.ApiGroup
	{
		groupRouter.GET("", baseApi.ListGroup)          // 获取组列表
		groupRouter.POST("", baseApi.CreateGroup)       // 创建组
		groupRouter.PUT("/:id", baseApi.UpdateGroup)    // 更新组
		groupRouter.DELETE("/:id", baseApi.DeleteGroup) // 删除组
	}
}

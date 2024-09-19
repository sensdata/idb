package router

import (
	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/core/api/entry"
	"github.com/sensdata/idb/center/core/api/middleware"
)

type UserRouter struct{}

func (s *UserRouter) InitRouter(Router *gin.RouterGroup) {
	userRouter := Router.Group("user")
	userRouter.Use(middleware.NewJWT().JWTAuth())
	baseApi := entry.ApiGroup
	{
		userRouter.POST("/list", baseApi.ListUser)
		userRouter.POST("/create", baseApi.CreateUser)
		userRouter.POST("/update", baseApi.UpdateUser)
		userRouter.POST("/delete", baseApi.DeleteUser)
		userRouter.POST("/valid", baseApi.ValidUser)
		userRouter.POST("/update/password", baseApi.ChangePassword)
	}
}

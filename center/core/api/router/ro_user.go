package router

import (
	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/core/api/entry"
	"github.com/sensdata/idb/center/core/api/middleware"
)

type UserRouter struct{}

func (s *UserRouter) InitRouter(Router *gin.RouterGroup) {
	userRouter := Router.Group("users")
	userRouter.Use(middleware.NewJWT().JWTAuth())
	baseApi := entry.ApiGroup
	{
		userRouter.GET("", baseApi.ListUser)                // 获取用户列表
		userRouter.POST("", baseApi.CreateUser)             // 新增用户
		userRouter.PUT("", baseApi.UpdateUser)              // 更新用户
		userRouter.DELETE("", baseApi.DeleteUser)           // 删除用户
		userRouter.PUT("/valid", baseApi.ValidUser)         // 禁用/启用用户
		userRouter.PUT("/password", baseApi.ChangePassword) // 更新密码
	}
}

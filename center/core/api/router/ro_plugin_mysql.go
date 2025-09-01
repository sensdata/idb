package router

import (
	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/core/api/entry"
	"github.com/sensdata/idb/center/core/api/middleware"
)

type MysqlRouter struct{}

func (s *MysqlRouter) InitRouter(Router *gin.RouterGroup) {
	mysqlRouter := Router.Group("mysql")
	mysqlRouter.Use(middleware.NewJWT().JWTAuth())
	baseApi := entry.ApiGroup
	{
		mysqlRouter.GET("/:host", baseApi.MysqlComposes)
		mysqlRouter.POST("/:host/operation", baseApi.MysqlOperation)
		mysqlRouter.POST("/:host/port", baseApi.MysqlSetPort)
		mysqlRouter.GET("/:host/conf", baseApi.MysqlGetConf)
		mysqlRouter.POST("/:host/conf", baseApi.MysqlSetConf)
		mysqlRouter.GET("/:host/remote_access", baseApi.MysqlGetRemoteAccess)
		mysqlRouter.POST("/:host/remote_access", baseApi.MysqlSetRemoteAccess)
		mysqlRouter.GET("/:host/password", baseApi.MysqlGetRootPassword)
		mysqlRouter.POST("/:host/password", baseApi.MysqlSetRootPassword)
	}
}

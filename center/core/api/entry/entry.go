// @title IDB API
// @version 1.0
// @description IDB apis doc genarated by swaggo
// @BasePath /api/v1
// @host http://8.138.47.21:9918
// @schemes http

package entry

import "github.com/sensdata/idb/center/core/api/service"

type BaseApi struct{}

var ApiGroup = new(BaseApi)

var (
	authService    = service.NewIAuthService()
	userService    = service.NewIUserService()
	groupService   = service.NewIGroupService()
	hostService    = service.NewIHostService()
	commandService = service.NewICommandService()
	actionService  = service.NewIActionService()
)

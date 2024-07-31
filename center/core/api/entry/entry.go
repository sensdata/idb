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
)

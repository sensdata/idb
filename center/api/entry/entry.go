package entry

import "github.com/sensdata/idb/center/api/service"

type BaseApi struct{}

var ApiGroup = new(BaseApi)

var (
	authService = service.NewIAuthService()
	userService = service.NewIUserService()
)

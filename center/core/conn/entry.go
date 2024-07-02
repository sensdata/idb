package conn

import (
	"github.com/sensdata/idb/center/config"
	"github.com/sensdata/idb/center/db/repo"
)

var (
	CommonRepo = repo.NewCommonRepo()

	RoleRepo      = repo.NewRoleRepo()
	UserRepo      = repo.NewUserRepo()
	GroupRepo     = repo.NewGroupRepo()
	HostRepo      = repo.NewHostRepo()
	HostGroupRepo = repo.NewHostGroupRepo()

	CONFMAN *config.Manager
	SSH     ISSHService
	CENTER  ICenter
)

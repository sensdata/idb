package service

import (
	"github.com/sensdata/idb/center/db/repo"
)

var (
	CommonRepo = repo.NewCommonRepo()

	RoleRepo       = repo.NewRoleRepo()
	UserRepo       = repo.NewUserRepo()
	GroupRepo      = repo.NewGroupRepo()
	HostRepo       = repo.NewHostRepo()
	HostGroupRepo  = repo.NewHostGroupRepo()
	AppRepo        = repo.NewAppRepo()
	AppVersionRepo = repo.NewAppVersionRepo()
)

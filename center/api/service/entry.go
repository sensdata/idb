package service

import "github.com/sensdata/idb/center/db"

var (
	CommonRepo = db.NewCommonRepo()

	RoleRepo = db.NewRoleRepo()
	UserRepo = db.NewUserRepo()
)

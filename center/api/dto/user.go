package dto

import "github.com/sensdata/idb/center/db/model"

type ListUser struct {
	PageInfo
}

type ListUserResult struct {
	Users []model.User
}

package db

import (
	"github.com/sensdata/idb/center/db/model"
	"github.com/sensdata/idb/center/global"
)

type RoleRepo struct{}

type IRoleRepo interface {
	Get(opts ...DBOption) (model.Role, error)
	GetList(opts ...DBOption) ([]model.Role, error)
}

func NewRoleRepo() IRoleRepo {
	return &RoleRepo{}
}

func (r *RoleRepo) Get(opts ...DBOption) (model.Role, error) {
	var role model.Role
	db := global.DB.Model(&model.Role{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&role).Error
	return role, err
}

func (r *RoleRepo) GetList(opts ...DBOption) ([]model.Role, error) {
	var roles []model.Role
	db := global.DB.Model(&model.Role{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&roles).Error
	return roles, err
}

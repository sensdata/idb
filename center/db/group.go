package db

import (
	"github.com/sensdata/idb/center/db/model"
	"github.com/sensdata/idb/center/global"
	"gorm.io/gorm"
)

type GroupRepo struct{}

type IGroupRepo interface {
	Get(opts ...DBOption) (model.Group, error)
	GetList(opts ...DBOption) ([]model.Group, error)
	Page(page, offset int, opts ...DBOption) (int64, []model.Group, error)
	Create(group *model.Group) error
	Update(id uint, vars map[string]interface{}) error
	Delete(opts ...DBOption) error
	WithByName(name string) DBOption
}

func NewGroupRepo() IGroupRepo {
	return &GroupRepo{}
}

func (r *GroupRepo) Get(opts ...DBOption) (model.Group, error) {
	var group model.Group
	db := global.DB.Model(&model.Group{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&group).Error
	return group, err
}

func (r *GroupRepo) GetList(opts ...DBOption) ([]model.Group, error) {
	var groups []model.Group
	db := global.DB.Model(&model.Group{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&groups).Error
	return groups, err
}

func (r *GroupRepo) Page(page, size int, opts ...DBOption) (int64, []model.Group, error) {
	var groups []model.Group
	db := global.DB.Model(&model.Group{})
	for _, opt := range opts {
		db = opt(db)
	}
	count := int64(0)
	db = db.Count(&count)
	err := db.Limit(size).Offset(size * (page - 1)).Find(&groups).Error
	return count, groups, err
}

func (r *GroupRepo) Create(group *model.Group) error {
	return global.DB.Create(group).Error
}

func (r *GroupRepo) Update(id uint, vars map[string]interface{}) error {
	return global.DB.Model(&model.Group{}).Where("id = ?", id).Updates(vars).Error
}

func (r *GroupRepo) Delete(opts ...DBOption) error {
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Delete(&model.Group{}).Error
}

func (r *GroupRepo) WithByName(name string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("group_name = ?", name)
	}
}

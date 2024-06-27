package repo

import (
	"github.com/sensdata/idb/center/db/model"
	"github.com/sensdata/idb/center/global"
	"gorm.io/gorm"
)

type HostGroupRepo struct{}

type IHostGroupRepo interface {
	Get(opts ...DBOption) (model.HostGroup, error)
	GetList(opts ...DBOption) ([]model.HostGroup, error)
	Page(page, size int, opts ...DBOption) (int64, []model.HostGroup, error)
	Create(group *model.HostGroup) error
	Update(id uint, vars map[string]interface{}) error
	Delete(opts ...DBOption) error
	WithByName(name string) DBOption
	WithById(id uint) DBOption
}

func NewHostGroupRepo() IHostGroupRepo {
	return &HostGroupRepo{}
}

func (r *HostGroupRepo) Get(opts ...DBOption) (model.HostGroup, error) {
	var group model.HostGroup
	db := global.DB.Model(&model.HostGroup{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&group).Error
	return group, err
}

func (r *HostGroupRepo) GetList(opts ...DBOption) ([]model.HostGroup, error) {
	var groups []model.HostGroup
	db := global.DB.Model(&model.HostGroup{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&groups).Error
	return groups, err
}

func (r *HostGroupRepo) Page(page, size int, opts ...DBOption) (int64, []model.HostGroup, error) {
	var groups []model.HostGroup
	db := global.DB.Model(&model.HostGroup{})
	for _, opt := range opts {
		db = opt(db)
	}
	count := int64(0)
	db = db.Count(&count)
	err := db.Limit(size).Offset(size * (page - 1)).Find(&groups).Error
	return count, groups, err
}

func (r *HostGroupRepo) Create(group *model.HostGroup) error {
	return global.DB.Create(group).Error
}

func (r *HostGroupRepo) Update(id uint, vars map[string]interface{}) error {
	return global.DB.Model(&model.HostGroup{}).Where("id = ?", id).Updates(vars).Error
}

func (r *HostGroupRepo) Delete(opts ...DBOption) error {
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Delete(&model.HostGroup{}).Error
}

func (r *HostGroupRepo) WithByName(name string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("group_name = ?", name)
	}
}

func (r *HostGroupRepo) WithByID(id uint) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("id = ?", id)
	}
}

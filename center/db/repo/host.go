package repo

import (
	"github.com/sensdata/idb/center/db/model"
	"github.com/sensdata/idb/center/global"
	"gorm.io/gorm"
)

type HostRepo struct{}

type IHostRepo interface {
	Get(opts ...DBOption) (model.Host, error)
	GetList(opts ...DBOption) ([]model.Host, error)
	Page(page, size int, opts ...DBOption) (int64, []model.Host, error)
	WithByID(id uint) DBOption
	WithByDefault() DBOption
	WithByName(name string) DBOption
	WithByAddr(addr string) DBOption
	WithByGroupID(groupID uint) DBOption
	Create(host *model.Host) error
	Update(id uint, vars map[string]interface{}) error
	Delete(opts ...DBOption) error
}

func NewHostRepo() IHostRepo {
	return &HostRepo{}
}

func (r *HostRepo) Get(opts ...DBOption) (model.Host, error) {
	var host model.Host
	db := global.DB.Model(&model.Host{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&host).Error
	return host, err
}

func (r *HostRepo) GetList(opts ...DBOption) ([]model.Host, error) {
	var hosts []model.Host
	db := global.DB.Model(&model.Host{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&hosts).Error
	return hosts, err
}

func (r *HostRepo) Page(page, size int, opts ...DBOption) (int64, []model.Host, error) {
	var hosts []model.Host
	db := global.DB.Model(&model.Host{})
	for _, opt := range opts {
		db = opt(db)
	}
	count := int64(0)
	db = db.Count(&count)
	err := db.Limit(size).Offset(size * (page - 1)).Find(&hosts).Error
	return count, hosts, err
}

func (r *HostRepo) WithByID(id uint) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("id = ?", id)
	}
}

func (r *HostRepo) WithByDefault() DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("default = ?", true)
	}
}

func (r *HostRepo) WithByName(name string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("name = ?", name)
	}
}

func (r *HostRepo) WithByAddr(addr string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("addr = ?", addr)
	}
}

func (c *HostRepo) WithByGroupID(groupID uint) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		if groupID == 0 {
			return g
		}
		return g.Where("group_id = ?", groupID)
	}
}

func (r *HostRepo) Create(user *model.Host) error {
	return global.DB.Create(user).Error
}

func (h *HostRepo) Update(id uint, vars map[string]interface{}) error {
	return global.DB.Model(&model.Host{}).Where("id = ?", id).Updates(vars).Error
}

func (h *HostRepo) Delete(opts ...DBOption) error {
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Delete(&model.Host{}).Error
}

package repo

import (
	"github.com/sensdata/idb/center/db/model"
	"github.com/sensdata/idb/center/global"
	"gorm.io/gorm"
)

type AppVersionRepo struct{}

type IAppVersionRepo interface {
	Get(opts ...DBOption) (model.AppVersion, error)
	GetList(opts ...DBOption) ([]model.AppVersion, error)
	Page(page, size int, opts ...DBOption) (int64, []model.AppVersion, error)
	WithByID(id uint) DBOption
	WithByAppID(appID uint) DBOption
	Create(app *model.AppVersion) error
	Update(id uint, vars map[string]interface{}) error
	Delete(opts ...DBOption) error
}

func NewAppVersionRepo() IAppVersionRepo {
	return &AppVersionRepo{}
}

func (r *AppVersionRepo) Get(opts ...DBOption) (model.AppVersion, error) {
	var app model.AppVersion
	db := global.DB.Model(&model.AppVersion{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&app).Error
	return app, err
}
func (r *AppVersionRepo) GetList(opts ...DBOption) ([]model.AppVersion, error) {
	var apps []model.AppVersion
	db := global.DB.Model(&model.AppVersion{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&apps).Error
	return apps, err
}
func (r *AppVersionRepo) Page(page, size int, opts ...DBOption) (int64, []model.AppVersion, error) {
	var apps []model.AppVersion
	db := global.DB.Model(&model.AppVersion{})
	for _, opt := range opts {
		db = opt(db)
	}
	count := int64(0)
	db = db.Count(&count)
	err := db.Limit(size).Offset(size * (page - 1)).Find(&apps).Error
	return count, apps, err
}
func (r *AppVersionRepo) WithByID(id uint) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("id = ?", id)
	}
}
func (r *AppVersionRepo) WithByAppID(appID uint) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("app_id = ?", appID)
	}
}
func (r *AppVersionRepo) Create(app *model.AppVersion) error {
	return global.DB.Create(app).Error
}
func (r *AppVersionRepo) Update(id uint, vars map[string]interface{}) error {
	return global.DB.Model(&model.AppVersion{}).Where("id = ?", id).Updates(vars).Error
}
func (r *AppVersionRepo) Delete(opts ...DBOption) error {
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Delete(&model.AppVersion{}).Error
}

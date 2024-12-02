package repo

import (
	"github.com/sensdata/idb/center/db/model"
	"github.com/sensdata/idb/center/global"
	"gorm.io/gorm"
)

type AppRepo struct{}

type IAppRepo interface {
	Get(opts ...DBOption) (model.App, error)
	GetList(opts ...DBOption) ([]model.App, error)
	Page(page, size int, opts ...DBOption) (int64, []model.App, error)
	WithByID(id uint) DBOption
	WithByName(name string) DBOption
	WithByCategory(category string) DBOption
	Create(app *model.App) error
	Update(id uint, vars map[string]interface{}) error
	Delete(opts ...DBOption) error
}

func NewAppRepo() IAppRepo {
	return &AppRepo{}
}

func (r *AppRepo) Get(opts ...DBOption) (model.App, error) {
	var app model.App
	db := global.DB.Model(&model.App{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&app).Error
	return app, err
}
func (r *AppRepo) GetList(opts ...DBOption) ([]model.App, error) {
	var apps []model.App
	db := global.DB.Model(&model.App{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&apps).Error
	return apps, err
}
func (r *AppRepo) Page(page, size int, opts ...DBOption) (int64, []model.App, error) {
	var apps []model.App
	db := global.DB.Model(&model.App{})
	for _, opt := range opts {
		db = opt(db)
	}
	count := int64(0)
	db = db.Count(&count)
	err := db.Limit(size).Offset(size * (page - 1)).Find(&apps).Error
	return count, apps, err
}
func (r *AppRepo) WithByID(id uint) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("id = ?", id)
	}
}
func (r *AppRepo) WithByName(name string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("name = ?", name)
	}
}
func (r *AppRepo) WithByCategory(category string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("category = ?", category)
	}
}
func (r *AppRepo) Create(app *model.App) error {
	return global.DB.Create(app).Error
}
func (r *AppRepo) Update(id uint, vars map[string]interface{}) error {
	return global.DB.Model(&model.App{}).Where("id = ?", id).Updates(vars).Error
}
func (r *AppRepo) Delete(opts ...DBOption) error {
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Delete(&model.App{}).Error
}

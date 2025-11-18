package repo

import (
	"github.com/sensdata/idb/center/db/model"
	"github.com/sensdata/idb/center/global"
	"gorm.io/gorm"
)

type SettingsRepo struct{}

type ISettingsRepo interface {
	GetList(opts ...DBOption) ([]model.Setting, error)
	Get(opts ...DBOption) (model.Setting, error)
	Create(key, value string) error
	Update(key, value string) error
	Upsert(key, value string) error
	WithByKey(key string) DBOption
}

func NewSettingsRepo() ISettingsRepo {
	return &SettingsRepo{}
}

func (u *SettingsRepo) GetList(opts ...DBOption) ([]model.Setting, error) {
	var settings []model.Setting
	db := global.DB.Model(&model.Setting{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&settings).Error
	return settings, err
}

func (u *SettingsRepo) Get(opts ...DBOption) (model.Setting, error) {
	var settings model.Setting
	db := global.DB.Model(&model.Setting{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&settings).Error
	return settings, err
}

func (u *SettingsRepo) Create(key, value string) error {
	setting := &model.Setting{
		Key:   key,
		Value: value,
	}
	return global.DB.Create(setting).Error
}

func (u *SettingsRepo) Update(key, value string) error {
	return global.DB.Model(&model.Setting{}).Where("key = ?", key).Updates(map[string]interface{}{"value": value}).Error
}

func (u *SettingsRepo) Upsert(key, value string) error {
	// 先尝试更新
	result := global.DB.Model(&model.Setting{}).
		Where("key = ?", key).
		Updates(map[string]interface{}{"value": value})

	if result.Error != nil {
		return result.Error
	}

	// 如果没有更新任何行，则执行插入
	if result.RowsAffected == 0 {
		return u.Create(key, value)
	}

	return nil
}

func (c *SettingsRepo) WithByKey(key string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("key = ?", key)
	}
}

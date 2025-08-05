package repo

import (
	"github.com/sensdata/idb/core/model"
	"gorm.io/gorm"
)

type FingerprintRepo struct{}

type IFingerprintRepo interface {
	GetFirst(opts ...DBOption) (*model.Fingerprint, error)
	WithByFingerprint(fingerprint string) DBOption
	Create(fingerprint *model.Fingerprint) error
	Update(id uint, vars map[string]interface{}) error
}

func NewFingerprintRepo() IFingerprintRepo {
	return &FingerprintRepo{}
}

func (f *FingerprintRepo) GetFirst(opts ...DBOption) (*model.Fingerprint, error) {
	var fingerprint model.Fingerprint
	db := getDb(opts...).Model(&model.Fingerprint{})
	if err := db.First(&fingerprint).Error; err != nil {
		return &fingerprint, err
	}
	return &fingerprint, nil
}

func (f *FingerprintRepo) WithByFingerprint(fingerprint string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("fingerprint = ?", fingerprint)
	}
}

func (f *FingerprintRepo) Create(fingerprint *model.Fingerprint) error {
	db := getDb().Model(&model.Fingerprint{})
	return db.Create(&fingerprint).Error
}

func (f *FingerprintRepo) Update(id uint, vars map[string]interface{}) error {
	db := getDb().Model(&model.Fingerprint{})
	return db.Where("id = ?", id).Updates(vars).Error
}

package repo

import (
	"github.com/sensdata/idb/center/db/model"
	"github.com/sensdata/idb/center/global"
)

type TimezonesRepo struct{}

type ITimezonesRepo interface {
	GetList(opts ...DBOption) ([]model.Timezone, error)
}

func NewTimezonesRepo() ITimezonesRepo {
	return &TimezonesRepo{}
}
func (t *TimezonesRepo) GetList(opts ...DBOption) ([]model.Timezone, error) {
	var timezones []model.Timezone
	db := global.DB.Model(&model.Timezone{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&timezones).Error
	return timezones, err
}

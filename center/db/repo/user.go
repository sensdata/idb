package repo

import (
	"github.com/sensdata/idb/center/db/model"
	"github.com/sensdata/idb/center/global"
	"gorm.io/gorm"
)

type UserRepo struct{}

type IUserRepo interface {
	Get(opts ...DBOption) (model.User, error)
	GetList(opts ...DBOption) ([]model.User, error)
	Page(page, size int, opts ...DBOption) (int64, []model.User, error)
	WithByID(id uint) DBOption
	WithByName(name string) DBOption
	Create(user *model.User) error
	Update(id uint, vars map[string]interface{}) error
	Delete(opts ...DBOption) error
}

func NewUserRepo() IUserRepo {
	return &UserRepo{}
}

func (r *UserRepo) Get(opts ...DBOption) (model.User, error) {
	var user model.User
	db := global.DB.Model(&model.User{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&user).Error
	return user, err
}

func (r *UserRepo) GetList(opts ...DBOption) ([]model.User, error) {
	var users []model.User
	db := global.DB.Model(&model.User{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&users).Error
	return users, err
}

func (r *UserRepo) Page(page, size int, opts ...DBOption) (int64, []model.User, error) {
	var users []model.User
	db := global.DB.Model(&model.User{})
	for _, opt := range opts {
		db = opt(db)
	}
	count := int64(0)
	db = db.Count(&count)
	err := db.Limit(size).Offset(size * (page - 1)).Find(&users).Error
	return count, users, err
}

func (r *UserRepo) WithByID(id uint) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("id = ?", id)
	}
}

func (r *UserRepo) WithByName(name string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("user_name = ?", name)
	}
}

func (r *UserRepo) Create(user *model.User) error {
	return global.DB.Create(user).Error
}

func (h *UserRepo) Update(id uint, vars map[string]interface{}) error {
	return global.DB.Model(&model.User{}).Where("id = ?", id).Updates(vars).Error
}

func (h *UserRepo) Delete(opts ...DBOption) error {
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Delete(&model.User{}).Error
}

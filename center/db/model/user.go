package model

// User 用户模型
type User struct {
	BaseModel

	Username string `gorm:"unique;not null" json:"userName"`
	Password string `gorm:"not null" json:"-"`
	Salt     string `gorm:"not null" json:"-"`
	RoleID   uint   `gorm:"not null" json:"-"`
}

package model

// User 用户模型
type User struct {
	BaseModel

	Name     string `gorm:"unique;not null" json:"name"`
	Password string `gorm:"not null" json:"-"`
	Salt     string `gorm:"not null" json:"-"`
	RoleID   uint   `gorm:"not null" json:"-"`
	GroupID  uint   `gorm:"not null" json:"-"`
	Valid    uint   `gorm:"not null" json:"valid"`
}

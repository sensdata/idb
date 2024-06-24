package db

import "time"

// Role 角色模型
type Role struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"unique;not null"`
	Description string
}

// User 用户模型
type User struct {
	ID        uint   `gorm:"primaryKey"`
	Username  string `gorm:"unique;not null"`
	Password  string `gorm:"not null"`
	Salt      string `gorm:"not null"`
	RoleID    uint   `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

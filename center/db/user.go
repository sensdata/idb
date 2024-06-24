package db

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// AddRole 添加新角色
func AddRole(db *gorm.DB, role Role) error {
	if err := db.Create(&role).Error; err != nil {
		return fmt.Errorf("failed to add role: %v", err)
	}
	return nil
}

// GetRoleByID 根据ID获取角色
func GetRoleByID(db *gorm.DB, id uint) (*Role, error) {
	var role Role
	if err := db.First(&role, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get role: %v", err)
	}
	return &role, nil
}

// AddUser 添加新用户
func AddUser(db *gorm.DB, user User) error {
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	if err := db.Create(&user).Error; err != nil {
		return fmt.Errorf("failed to add user: %v", err)
	}
	return nil
}

// GetUserByUsername 根据用户名获取用户
func GetUserByUsername(db *gorm.DB, username string) (*User, error) {
	var user User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user: %v", err)
	}
	return &user, nil
}

// UpdateUser 更新用户信息
func UpdateUser(db *gorm.DB, user User) error {
	user.UpdatedAt = time.Now()
	if err := db.Save(&user).Error; err != nil {
		return fmt.Errorf("failed to update user: %v", err)
	}
	return nil
}

// DeleteUser 删除用户
func DeleteUser(db *gorm.DB, id uint) error {
	if err := db.Delete(&User{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete user: %v", err)
	}
	return nil
}

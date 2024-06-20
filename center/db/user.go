package db

import (
	"database/sql"
	"fmt"
	"time"
)

// AddRole 添加新角色
func AddRole(role Role) error {
	_, err := db.Exec("INSERT INTO roles (name, description) VALUES (?, ?)", role.Name, role.Description)
	if err != nil {
		return fmt.Errorf("failed to add role: %v", err)
	}
	return nil
}

// GetRoleByID 根据ID获取角色
func GetRoleByID(id int) (*Role, error) {
	row := db.QueryRow("SELECT id, name, description FROM roles WHERE id = ?", id)
	var role Role
	if err := row.Scan(&role.ID, &role.Name, &role.Description); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get role: %v", err)
	}
	return &role, nil
}

// AddUser 添加新用户
func AddUser(user User) error {
	_, err := db.Exec("INSERT INTO users (username, password, salt, role_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
		user.Username, user.Password, user.RoleID, time.Now(), time.Now())
	if err != nil {
		return fmt.Errorf("failed to add user: %v", err)
	}
	return nil
}

// GetUserByUsername 根据用户名获取用户
func GetUserByUsername(username string) (*User, error) {
	row := db.QueryRow("SELECT id, username, password, salt, role_id, created_at, updated_at FROM users WHERE username = ?", username)
	var user User
	if err := row.Scan(&user.ID, &user.Username, &user.Password, &user.RoleID, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user: %v", err)
	}
	return &user, nil
}

// UpdateUser 更新用户信息
func UpdateUser(user User) error {
	_, err := db.Exec("UPDATE users SET username = ?, password = ?, salt = ?, role_id = ?, updated_at = ? WHERE id = ?",
		user.Username, user.Password, user.Salt, user.RoleID, time.Now(), user.ID)
	if err != nil {
		return fmt.Errorf("failed to update user: %v", err)
	}
	return nil
}

// DeleteUser 删除用户
func DeleteUser(id int) error {
	_, err := db.Exec("DELETE FROM users WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %v", err)
	}
	return nil
}

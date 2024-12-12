package model

import "time"

type RoleInfo struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	RoleName  string    `json:"role_name"`
}

type UserInfo struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	RoleInfo  RoleInfo  `json:"role"`
	GroupInfo GroupInfo `json:"group"`
	Valid     uint      `json:"valid"`
}

type CreateUser struct {
	Name     string `json:"name" validate:"required"`
	Password string `json:"password" validate:"required"`
	GroupID  uint   `json:"group_id" validate:"required"`
}

type UpdateUser struct {
	ID      uint   `json:"id"`
	Name    string `json:"name" validate:"required"`
	GroupID uint   `json:"group_id" validate:"required"`
	Valid   uint   `json:"valid" validate:"required"`
}

type ValidUser struct {
	ID    uint `json:"id"`
	Valid uint `json:"valid"`
}

type ChangePassword struct {
	ID       uint   `json:"id"`
	Password string `json:"password" validate:"required"`
}

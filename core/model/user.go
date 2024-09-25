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
	UserName  string    `json:"user_name"`
	RoleInfo  RoleInfo  `json:"role"`
	GroupInfo GroupInfo `json:"group"`
	Valid     uint      `json:"valid"`
}

type CreateUser struct {
	UserName string `json:"user_name" validate:"required"`
	Password string `json:"password" validate:"required"`
	GroupID  uint   `json:"group_id" validate:"required,number"`
}

type UpdateUser struct {
	UserName string `json:"user_name" validate:"required"`
	GroupID  uint   `json:"group_id" validate:"required,number"`
	Valid    uint   `json:"valid" validate:"required,number"`
}

type ValidUser struct {
	Valid uint `json:"valid" validate:"required,number"`
}

type ChangePassword struct {
	Password string `json:"password" validate:"required"`
}

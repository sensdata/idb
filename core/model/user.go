package model

import "time"

type RoleInfo struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	RoleName  string    `json:"roleName"`
}

type UserInfo struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UserName  string    `json:"userName"`
	RoleInfo  RoleInfo  `json:"role"`
	GroupInfo GroupInfo `json:"group"`
	Valid     uint      `json:"valid"`
}

type CreateUser struct {
	UserName string `json:"userName" validate:"required"`
	Password string `json:"password" validate:"required"`
	GroupID  uint   `json:"groupId" validate:"required,number"`
}

type UpdateUser struct {
	UserName string `json:"userName" validate:"required"`
	GroupID  uint   `json:"groupId" validate:"required,number"`
	Valid    uint   `json:"valid" validate:"required,number"`
}

type ValidUser struct {
	Valid uint `json:"valid" validate:"required,number"`
}

type ChangePassword struct {
	Password string `json:"password" validate:"required"`
}

package model

import "time"

type GroupInfo struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	GroupName string    `json:"groupName"`
}

type CreateGroup struct {
	GroupName string `json:"groupName" validate:"required"`
}

type UpdateGroup struct {
	GroupID   uint   `json:"groupId" validate:"required,number"`
	GroupName string `json:"groupName" validate:"required"`
}

type DeleteGroup struct {
	GroupID uint `json:"groupId" validate:"required,number"`
}

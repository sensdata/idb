package model

import "time"

type GroupInfo struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	GroupName string    `json:"group_name"`
}

type CreateGroup struct {
	GroupName string `json:"group_name" validate:"required"`
}

type UpdateGroup struct {
	GroupName string `json:"group_name" validate:"required"`
}

type DeleteGroup struct {
	GroupID uint `json:"group_id" validate:"required,number"`
}

package model

// Role 角色模型
type Role struct {
	BaseModel

	Name        string `gorm:"unique;not null" json:"name"`
	Description string `json:"description"`
}

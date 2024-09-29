package model

// Group 分组
type Group struct {
	BaseModel

	GroupName string `gorm:"unique;not null" json:"group_name"`
}

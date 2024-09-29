package model

import "time"

type BaseModel struct {
	ID        uint      `gorm:"primarykey;AUTO_INCREMENT" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

package model

type Favorite struct {
	BaseModel
	Name  string `gorm:"type:varchar(256);not null;" json:"name" `
	Path  string `gorm:"type:varchar(256);not null;unique" json:"path"`
	Type  string `gorm:"type:varchar(64);" json:"type"`
	IsDir bool   `json:"isDir"`
	IsTxt bool   `json:"isTxt"`
}

type FavoriteCreate struct {
	Path string `json:"path" validate:"required"`
}

type FavoriteDelete struct {
	ID uint `json:"id" validate:"required"`
}

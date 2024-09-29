package model

type Favorite struct {
	BaseModel
	Name   string `gorm:"type:varchar(256);not null;" json:"name" `
	Source string `gorm:"type:varchar(256);not null;unique" json:"source"`
	Type   string `gorm:"type:varchar(64);" json:"type"`
	IsDir  bool   `json:"is_dir"`
	IsTxt  bool   `json:"is_txt"`
}

type FavoriteListReq struct {
	PageInfo
	HostID uint `json:"host_id"`
}

type FavoriteCreate struct {
	HostID uint   `json:"host_id"`
	Source string `json:"source" validate:"required"`
}

type FavoriteDelete struct {
	HostID uint `json:"host_id"`
	ID     uint `json:"id" validate:"required"`
}

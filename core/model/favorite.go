package model

type Favorite struct {
	BaseModel
	Name  string `gorm:"type:varchar(256);not null;" json:"name" `
	Path  string `gorm:"type:varchar(256);not null;unique" json:"path"`
	Type  string `gorm:"type:varchar(64);" json:"type"`
	IsDir bool   `json:"isDir"`
	IsTxt bool   `json:"isTxt"`
}

type FavoriteListReq struct {
	PageInfo
	HostID uint `json:"hostId"`
}

type FavoriteCreate struct {
	HostID uint   `json:"hostId"`
	Path   string `json:"path" validate:"required"`
}

type FavoriteDelete struct {
	HostID uint `json:"hostId"`
	ID     uint `json:"id" validate:"required"`
}

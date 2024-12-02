package model

type AppVersion struct {
	BaseModel

	AppId          uint   `gorm:"type:integer;not null" json:"-"`
	Version        string `gorm:"type:varchar(64);not null" json:"-"`
	FormContent    string `gorm:"type:longtext;not null" json:"-"`
	ComposeContent string `gorm:"type:longtext;not null" json:"-"`
	EnvContent     string `gorm:"type:longtext;not null" json:"-"`
	ConfigName     string `gorm:"type:varchar(128);not null" json:"-"`
	ConfigContent  string `gorm:"type:longtext;not null" json:"-"`
}

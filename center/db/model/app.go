package model

type App struct {
	BaseModel

	Name        string `gorm:"unique;type:varchar(64);not null" json:"-"`
	DisplayName string `gorm:"type:varchar(64);not null" json:"-"`
	Category    string `gorm:"type:varchar(64);not null" json:"-"`
	Tags        string `gorm:"type:longtext;not null" json:"-"`
	Title       string `gorm:"type:varchar(128);not null" json:"-"`
	Description string `gorm:"type:longtext;not null" json:"-"`
	Vendor      string `gorm:"type:varchar(128);not null" json:"-"`
	VendorUrl   string `gorm:"type:longtext;not null" json:"-"`
	Packager    string `gorm:"type:varchar(128);not null" json:"-"`
	PackagerUrl string `gorm:"type:longtext;not null" json:"-"`
}

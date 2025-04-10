package model

type Timezone struct {
	BaseModel
	Value  string `json:"value" gorm:"type:varchar(256)"` // 时区值标识
	Abbr   string `json:"abbr" gorm:"type:varchar(64)"`   // 时区缩写
	Offset int    `json:"offset" gorm:"type:int"`         // UTC偏移量(小时)
	IsDst  bool   `json:"isdst" gorm:"type:bool"`         // 是否为夏令时
	Text   string `json:"text" gorm:"type:varchar(256)"`  // 时区显示文本
	UTC    string `json:"utc" gorm:"type:varchar(256)"`   // UTC标准时区名称
}

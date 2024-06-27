package model

// Host group 主机分组
type HostGroup struct {
	BaseModel

	GroupName string `gorm:"unique;not null" json:"groupName"`
}

// Host 主机模型
type Host struct {
	BaseModel

	GroupID    uint   `gorm:"type:decimal;not null" json:"groupID"`
	Name       string `gorm:"type:varchar(64);not null" json:"name"`
	Addr       string `gorm:"type:varchar(16);not null" json:"addr"`
	Port       int    `gorm:"type:decimal;not null" json:"port"`
	User       string `gorm:"type:varchar(64);not null" json:"user"`
	AuthMode   string `gorm:"type:varchar(16);not null" json:"authMode"`
	Password   string `gorm:"type:varchar(64)" json:"password"`
	PrivateKey string `gorm:"type:varchar(256)" json:"privateKey"`
	PassPhrase string `gorm:"type:varchar(256)" json:"passPhrase"`

	AgentAddr string `gorm:"type:varchar(16);not null" json:"agentAddr"`
	AgentPort int    `gorm:"type:decimal;not null" json:"agentPort"`
	AgentMode string `gorm:"type:varchar(16);not null" json:"agentMode"`
	AgentKey  string `gorm:"type:varchar(32);not null" json:"agentKey"`
}

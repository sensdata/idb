package model

// Host group 主机分组
type HostGroup struct {
	BaseModel

	GroupName string `gorm:"unique;not null" json:"group_name"`
}

// Host 主机模型
type Host struct {
	BaseModel

	GroupID    uint   `gorm:"type:decimal;not null" json:"group_id"`
	Name       string `gorm:"type:varchar(64);not null" json:"name"`
	Addr       string `gorm:"type:varchar(16);not null" json:"addr"`
	Port       int    `gorm:"type:decimal;not null" json:"port"`
	User       string `gorm:"type:varchar(64);not null" json:"user"`
	AuthMode   string `gorm:"type:varchar(16);not null" json:"auth_mode"`
	Password   string `gorm:"type:varchar(64)" json:"password"`
	PrivateKey string `gorm:"type:varchar(256)" json:"private_key"`
	PassPhrase string `gorm:"type:varchar(256)" json:"pass_phrase"`

	AgentAddr string `gorm:"type:varchar(16);not null" json:"agent_addr"`
	AgentPort int    `gorm:"type:decimal;not null" json:"agent_port"`
	AgentMode string `gorm:"type:varchar(16);not null" json:"agent_mode"`
	AgentKey  string `gorm:"type:varchar(32);not null" json:"agent_key"`
}

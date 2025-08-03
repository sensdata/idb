package model

import "time"

type HostInfo struct {
	ID         uint      `json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	Default    bool      `json:"default"`
	Serial     string    `json:"serial"`
	GroupInfo  GroupInfo `json:"group"`
	Name       string    `json:"name"`
	Addr       string    `json:"addr"`
	Port       int       `json:"port"`
	User       string    `json:"user"`
	AuthMode   string    `json:"auth_mode"`
	Password   string    `json:"password"`
	PrivateKey string    `json:"private_key"`
	PassPhrase string    `json:"pass_phrase"`

	AgentAddr    string      `json:"agent_addr"`
	AgentPort    int         `json:"agent_port"`
	AgentMode    string      `json:"agent_mode"`
	AgentKey     string      `json:"agent_key"`
	AgentVersion string      `json:"agent_version"`
	AgentStatus  AgentStatus `json:"agent_status"`
	AgentLatest  string      `json:"agent_latest"`
}

type ListHost struct {
	PageInfo
	GroupID uint   `form:"group_id"`
	Keyword string `form:"keyword"`
}

type CreateHost struct {
	GroupID    uint   `json:"group_id" validate:"required"`
	Name       string `json:"name" validate:"required"`
	Addr       string `json:"addr" validate:"required"`
	Port       int    `json:"port" validate:"required"`
	User       string `json:"user" validate:"required"`
	AuthMode   string `json:"auth_mode" validate:"required"`
	Password   string `json:"password"`
	PrivateKey string `json:"private_key"`
	PassPhrase string `json:"pass_phrase"`
}

type UpdateHost struct {
	GroupID uint   `json:"group_id" validate:"required"`
	ID      uint   `json:"id"`
	Name    string `json:"name" validate:"required"`
}

// Auth related models

type IssueLicenseReq struct {
	IP string `json:"ip" validate:"required,ip"`
}

type IssueLicenseResp struct {
	Serial string `json:"serial"`
}

type ReissueLicenseReq struct {
	OldIP     string `json:"old_ip" validate:"required,ip"`
	NewIP     string `json:"new_ip" validate:"required,ip"`
	OldSerial string `json:"old_serial" validate:"required"`
}

type BindLicenseReq struct {
	IP     string `json:"ip" validate:"required,ip"`
	Serial string `json:"serial" validate:"required"`
}

type VerifyLicenseReq struct {
	IP     string `json:"ip" validate:"required,ip"`
	Serial string `json:"serial" validate:"required"`
}

type UpdateHostSSH struct {
	Addr       string `json:"addr" validate:"required"`
	Port       int    `json:"port" validate:"required"`
	User       string `json:"user" validate:"required"`
	AuthMode   string `json:"auth_mode" validate:"required"`
	Password   string `json:"password"`
	PrivateKey string `json:"private_key"`
	PassPhrase string `json:"pass_phrase"`
}

type UpdateHostAgent struct {
	AgentAddr string `json:"agent_addr" validate:"required"`
	AgentPort int    `json:"agent_port" validate:"required"`
	AgentMode string `json:"agent_mode" validate:"required"`
	AgentKey  string `json:"agent_key"`
}

type TestSSH struct {
	Addr       string `json:"addr" validate:"required"`
	Port       int    `json:"port" validate:"required"`
	User       string `json:"user" validate:"required"`
	AuthMode   string `json:"auth_mode" validate:"required"`
	Password   string `json:"password"`
	PrivateKey string `json:"private_key"`
	PassPhrase string `json:"pass_phrase"`
}

type TestAgent struct {
	AgentAddr string `json:"agent_addr" validate:"required"`
	AgentPort int    `json:"agent_port" validate:"required"`
	AgentMode string `json:"agent_mode" validate:"required"`
	AgentKey  string `json:"agent_key"`
}

type InstallAgent struct {
	Upgrade bool `json:"upgrade"`
}

type AgentStatus struct {
	Status    string `json:"status"`
	Connected string `json:"connected"`
}

type HostStatus struct {
	Cpu      float64 `json:"cpu"`
	Memory   float64 `json:"mem"`
	MemTotal string  `json:"mem_total"` //总可用 = 物理内存 - 内核占用
	MemUsed  string  `json:"mem_used"`  //已使用 = 进程占用 + 缓冲区 + 缓存区
	Disk     float64 `json:"disk"`
	Rx       float64 `json:"rx"` //接收实时速率
	Tx       float64 `json:"tx"` //发送实时速率
}

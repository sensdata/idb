package model

import "time"

type HostInfo struct {
	ID         uint      `json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	Default    bool      `json:"default"`
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
	CanUpgrade   bool        `json:"can_upgrade"`
}

type ListHost struct {
	PageInfo
	GroupID uint   `form:"group_id" json:"group_id"`
	Keyword string `form:"keyword" json:"keyword"`
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
	Cpu       float64 `json:"cpu"`
	Memory    float64 `json:"mem"`
	MemTotal  string  `json:"mem_total"` //总可用 = 物理内存 - 内核占用
	MemUsed   string  `json:"mem_used"`  //已使用 = 进程占用 + 缓冲区 + 缓存区
	Disk      float64 `json:"disk"`
	Rx        float64 `json:"rx"`        //接收实时速率
	Tx        float64 `json:"tx"`        //发送实时速率
	Activated bool    `json:"activated"` // 是否已激活: verify_result > 0
}

type HostStatusInfo struct {
	Installed     string  `json:"installed"`
	Connected     string  `json:"connected"`
	Activated     bool    `json:"activated"`
	CanUpgrade    bool    `json:"can_upgrade"`
	Cpu           float64 `json:"cpu"`
	Memory        float64 `json:"mem"`
	MemTotal      string  `json:"mem_total"` //总可用 = 物理内存 - 内核占用
	MemUsed       string  `json:"mem_used"`  //已使用 = 进程占用 + 缓冲区 + 缓存区
	Disk          float64 `json:"disk"`
	Rx            float64 `json:"rx"` //接收实时速率
	Tx            float64 `json:"tx"` //发送实时速率
	LastHeartbeat int64   `json:"timestamp"`
}

func NewHostStatusInfo() *HostStatusInfo {
	return &HostStatusInfo{
		Installed:     "not installed",
		Connected:     "offline",
		Activated:     false,
		CanUpgrade:    false,
		Cpu:           0,
		Memory:        0,
		MemTotal:      "",
		MemUsed:       "",
		Disk:          0,
		Rx:            0,
		Tx:            0,
		LastHeartbeat: 0,
	}
}

type HostStatusDTO struct {
	ID         uint    `json:"id"`
	Installed  string  `json:"installed"`
	Connected  string  `json:"connected"`
	Activated  bool    `json:"activated"`
	CanUpgrade bool    `json:"can_upgrade"`
	Cpu        float64 `json:"cpu"`
	Memory     float64 `json:"mem"`
	MemTotal   string  `json:"mem_total"` //总可用 = 物理内存 - 内核占用
	MemUsed    string  `json:"mem_used"`  //已使用 = 进程占用 + 缓冲区 + 缓存区
	Disk       float64 `json:"disk"`
	Rx         float64 `json:"rx"` //接收实时速率
	Tx         float64 `json:"tx"` //发送实时速率
}

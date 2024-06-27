package dto

import "time"

type HostInfo struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"createdAt"`

	GroupInfo  GroupInfo `json:"group"`
	Name       string    `json:"name"`
	Addr       string    `json:"addr"`
	Port       int       `json:"port"`
	User       string    `json:"user"`
	AuthMode   string    `json:"authMode"`
	Password   string    `json:"password"`
	PrivateKey string    `json:"privateKey"`
	PassPhrase string    `json:"passPhrase"`

	AgentAddr string `json:"agentAddr"`
	AgentPort int    `json:"agentPort"`
	AgentMode string `json:"agentMode"`
	AgentKey  string `json:"agentKey"`
}

type ListHost struct {
	PageInfo
	GroupID uint   `json:"groupId" validate:"required,number"`
	Keyword string `json:"keyword"`
}

type CreateHost struct {
	GroupID    uint   `json:"groupId" validate:"required,number"`
	Name       string `json:"name" validate:"required"`
	Addr       string `json:"addr" validate:"required,string"`
	Port       int    `json:"port" validate:"required,number"`
	User       string `json:"user" validate:"required,string"`
	AuthMode   string `json:"authMode" validate:"required,string"`
	Password   string `json:"password"`
	PrivateKey string `json:"privateKey"`
	PassPhrase string `json:"passPhrase"`
}

type UpdateHost struct {
	HostID  uint   `json:"hostId" validate:"required,number"`
	GroupID uint   `json:"groupId" validate:"required,number"`
	Name    string `json:"name" validate:"required"`
}

type UpdateHostSSH struct {
	HostID     uint   `json:"hostId" validate:"required,number"`
	Addr       string `json:"addr" validate:"required,string"`
	Port       int    `json:"port" validate:"required,number"`
	User       string `json:"user" validate:"required,string"`
	AuthMode   string `json:"authMode" validate:"required,string"`
	Password   string `json:"password"`
	PrivateKey string `json:"privateKey"`
	PassPhrase string `json:"passPhrase"`
}

type UpdateHostAgent struct {
	HostID    uint   `json:"hostId" validate:"required,number"`
	AgentAddr string `json:"agentAddr" validate:"required,string"`
	AgentPort int    `json:"agentPort" validate:"required,number"`
	AgentMode string `json:"agentMode" validate:"required,string"`
	AgentKey  string `json:"agentKey"`
}

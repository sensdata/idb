package model

import "time"

type HostInfo struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`

	GroupInfo  GroupInfo `json:"group"`
	Name       string    `json:"name"`
	Addr       string    `json:"addr"`
	Port       int       `json:"port"`
	User       string    `json:"user"`
	AuthMode   string    `json:"auth_mode"`
	Password   string    `json:"password"`
	PrivateKey string    `json:"private_key"`
	PassPhrase string    `json:"pass_phrase"`

	AgentAddr string `json:"agent_addr"`
	AgentPort int    `json:"agent_port"`
	AgentMode string `json:"agent_mode"`
	AgentKey  string `json:"agent_key"`
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

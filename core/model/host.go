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
	GroupID uint   `json:"group_id"`
	Keyword string `json:"keyword"`
}

type CreateHost struct {
	GroupID    uint   `json:"group_id" validate:"required,number"`
	Name       string `json:"name" validate:"required"`
	Addr       string `json:"addr" validate:"required,string"`
	Port       int    `json:"port" validate:"required,number"`
	User       string `json:"user" validate:"required,string"`
	AuthMode   string `json:"auth_mode" validate:"required,string"`
	Password   string `json:"password"`
	PrivateKey string `json:"private_key"`
	PassPhrase string `json:"pass_phrase"`
}

type UpdateHost struct {
	GroupID uint   `json:"group_id" validate:"required,number"`
	Name    string `json:"name" validate:"required"`
}

type UpdateHostSSH struct {
	Addr       string `json:"addr" validate:"required,string"`
	Port       int    `json:"port" validate:"required,number"`
	User       string `json:"user" validate:"required,string"`
	AuthMode   string `json:"auth_mode" validate:"required,string"`
	Password   string `json:"password"`
	PrivateKey string `json:"private_key"`
	PassPhrase string `json:"pass_phrase"`
}

type UpdateHostAgent struct {
	AgentAddr string `json:"agent_addr" validate:"required,string"`
	AgentPort int    `json:"agent_port" validate:"required,number"`
	AgentMode string `json:"agent_mode" validate:"required,string"`
	AgentKey  string `json:"agent_key"`
}

type TestSSH struct {
	Addr       string `json:"addr" validate:"required,string"`
	Port       int    `json:"port" validate:"required,number"`
	User       string `json:"user" validate:"required,string"`
	AuthMode   string `json:"auth_mode" validate:"required,string"`
	Password   string `json:"password"`
	PrivateKey string `json:"private_key"`
	PassPhrase string `json:"pass_phrase"`
}

type TestAgent struct {
	AgentAddr string `json:"agent_addr" validate:"required,string"`
	AgentPort int    `json:"agent_port" validate:"required,number"`
	AgentMode string `json:"agent_mode" validate:"required,string"`
	AgentKey  string `json:"agent_key"`
}

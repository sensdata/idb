package model

import "time"

type SSHConfigReq struct {
	HostID uint `json:"host_id"`
}

type SSHInfo struct {
	AutoStart              bool   `json:"auto_start"`
	Status                 string `json:"status"`
	Message                string `json:"message"`
	Port                   string `json:"port"`
	ListenAddress          string `json:"listen_address"`
	PasswordAuthentication string `json:"password_authentication"`
	PubkeyAuthentication   string `json:"pubkey_authentication"`
	PermitRootLogin        string `json:"permit_root_login"`
	UseDNS                 string `json:"use_dns"`
}

type SSHUpdate struct {
	HostID uint                `json:"host_id"`
	Values []KeyValueForUpdate `json:"values"`
}

type SSHConfigContent struct {
	Content string `json:"content"`
}

type ContentUpdate struct {
	HostID  uint   `json:"host_id"`
	Content string `json:"content"`
}

type SSHOperate struct {
	HostID    uint   `json:"host_id"`
	Operation string `json:"operation"`
}

type GenerateKey struct {
	HostID         uint   `json:"host_id"`
	KeyBits        uint   `json:"key_bits" validate:"required,oneof=1024 2048"`
	Enabled        bool   `json:"enabled"`
	EncryptionMode string `json:"encryption_mode" validate:"required,oneof=rsa ed25519 ecdsa dsa"`
	Password       string `json:"password"`
}

type ListKey struct {
	HostID  uint   `json:"host_id"`
	Keyword string `json:"keyword"`
}

type KeyInfo struct {
	FileName    string
	Fingerprint string
	User        string
	Status      string
	KeyBits     int
}

type GenerateLoad struct {
	HostID         uint   `json:"host_id"`
	EncryptionMode string `json:"encryption_mode" validate:"required,oneof=rsa ed25519 ecdsa dsa"`
}

type SearchSSHLog struct {
	PageInfo
	HostID uint   `form:"host_id"`
	Info   string `form:"info"`
	Status string `form:"status"`
}
type SSHLog struct {
	Logs            []SSHHistory `json:"logs"`
	TotalCount      int          `json:"total_count"`
	SuccessfulCount int          `json:"successful_count"`
	FailedCount     int          `json:"failed_count"`
}

type SSHHistory struct {
	Date     time.Time `json:"date"`
	DateStr  string    `json:"date_str"`
	Area     string    `json:"area"`
	User     string    `json:"user"`
	AuthMode string    `json:"auth_mode"`
	Address  string    `json:"address"`
	Port     string    `json:"port"`
	Status   string    `json:"status"`
	Message  string    `json:"message"`
}

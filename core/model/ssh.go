package model

import "time"

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
	Values []KeyValueForUpdate `json:"values"`
}

type SSHConfigContent struct {
	Content string `json:"content"`
}

type ContentUpdate struct {
	Content string `json:"content"`
}

type SSHOperate struct {
	Operation string `json:"operation"`
}

type GenerateKey struct {
	KeyBits        uint   `json:"key_bits" validate:"required,oneof=1024 2048"`
	Enabled        bool   `json:"enabled"`
	EncryptionMode string `json:"encryption_mode" validate:"required,oneof=rsa ed25519 ecdsa dsa"`
	Password       string `json:"password"`
}

type ListKey struct {
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
	EncryptionMode string `json:"encryption_mode" validate:"required,oneof=rsa ed25519 ecdsa dsa"`
}

type SearchSSHLog struct {
	PageInfo
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

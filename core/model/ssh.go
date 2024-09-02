package model

import "time"

type SSHConfigReq struct {
	HostID uint `json:"hostId"`
}

type SSHInfo struct {
	AutoStart              bool   `json:"autoStart"`
	Status                 string `json:"status"`
	Message                string `json:"message"`
	Port                   string `json:"port"`
	ListenAddress          string `json:"listenAddress"`
	PasswordAuthentication string `json:"passwordAuthentication"`
	PubkeyAuthentication   string `json:"pubkeyAuthentication"`
	PermitRootLogin        string `json:"permitRootLogin"`
	UseDNS                 string `json:"useDNS"`
}

type SSHUpdate struct {
	HostID uint                `json:"hostId"`
	Values []KeyValueForUpdate `json:"values"`
}

type SSHConfigContent struct {
	Content string `json:"content"`
}

type ContentUpdate struct {
	HostID  uint   `json:"hostId"`
	Content string `json:"content"`
}

type SSHOperate struct {
	HostID    uint   `json:"hostId"`
	Operation string `json:"operation"`
}

type GenerateKey struct {
	HostID         uint   `json:"hostId"`
	KeyBits        uint   `json:"keyBits" validate:"required,oneof=1024 2048"`
	Enabled        bool   `json:"enabled"`
	EncryptionMode string `json:"encryptionMode" validate:"required,oneof=rsa ed25519 ecdsa dsa"`
	Password       string `json:"password"`
}

type ListKey struct {
	HostID  uint   `json:"hostId"`
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
	HostID         uint   `json:"hostId"`
	EncryptionMode string `json:"encryptionMode" validate:"required,oneof=rsa ed25519 ecdsa dsa"`
}

type SearchSSHLog struct {
	PageInfo
	HostID uint   `json:"hostId"`
	Info   string `json:"info"`
	Status string `json:"Status" validate:"required,oneof=Success Failed All"`
}
type SSHLog struct {
	Logs            []SSHHistory `json:"logs"`
	TotalCount      int          `json:"totalCount"`
	SuccessfulCount int          `json:"successfulCount"`
	FailedCount     int          `json:"failedCount"`
}

type SSHHistory struct {
	Date     time.Time `json:"date"`
	DateStr  string    `json:"dateStr"`
	Area     string    `json:"area"`
	User     string    `json:"user"`
	AuthMode string    `json:"authMode"`
	Address  string    `json:"address"`
	Port     string    `json:"port"`
	Status   string    `json:"status"`
	Message  string    `json:"message"`
}

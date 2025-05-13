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
	Operation string `json:"operation" validate:"required,oneof=enable disable stop reload restart"`
}

type GenerateKey struct {
	KeyName        string `json:"key_name" validate:"required"`
	KeyBits        uint   `json:"key_bits" validate:"required,oneof=1024 2048"`
	Enable         bool   `json:"enable"`
	EncryptionMode string `json:"encryption_mode" validate:"required,oneof=rsa ed25519 ecdsa dsa"`
	Password       string `json:"password"`
}

type ListKey struct {
	Keyword string `json:"keyword"`
}

type EnableKey struct {
	KeyName string `json:"key_name" validate:"required"`
	Enable  bool   `json:"enable"`
}

type RemoveKey struct {
	KeyName        string `json:"key_name" validate:"required"`
	OnlyPrivateKey bool   `json:"only_private_key" validate:"required"`
}

type SetKeyPassword struct {
	KeyName  string `json:"key_name" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type UpdateKeyPassword struct {
	KeyName     string `json:"key_name" validate:"required"`
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required"`
}

type KeyInfo struct {
	KeyName        string `json:"key_name"`
	Fingerprint    string `json:"fingerprint"`
	User           string `json:"user"`
	Status         string `json:"status"`
	KeyBits        int    `json:"key_bits"`
	HasPrivateKey  bool   `json:"has_private_key"`
	PrivateKeyPath string `json:"private_key_path"`
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

type AddAuthKey struct {
	Content string `json:"content" validate:"required"`
}

type RemoveAuthKey struct {
	Content string `json:"content" validate:"required"`
}

type AuthKeyInfo struct {
	Algorithm string `json:"algorithm"`
	Key       string `json:"key"`
	Comment   string `json:"comment"`
}

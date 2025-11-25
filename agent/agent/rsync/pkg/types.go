package pkg

import "time"

type SyncDirection string
type RemoteType string
type TaskState string
type AuthMode string

const (
	DirectionRemoteToLocal SyncDirection = "remote_to_local"
	DirectionLocalToRemote SyncDirection = "local_to_remote"

	RemoteTypeSSH   RemoteType = "ssh"
	RemoteTypeRsync RemoteType = "rsync"

	StatePending   TaskState = "pending"
	StateRunning   TaskState = "running"
	StateSucceeded TaskState = "succeeded"
	StateFailed    TaskState = "failed"
	StateStopped   TaskState = "stopped"

	AuthModePassword AuthMode = "password"
	AuthModeAnonymous AuthMode = "anonymous"
	AuthModePrivateKey AuthMode = "private_key"
)

// RsyncTask is the core persisted task model
type RsyncTask struct {
	ID            string        `json:"id"`
	Name          string        `json:"name"`
	Direction     SyncDirection `json:"direction"`
	LocalPath     string        `json:"local_path"`
	RemoteType    RemoteType    `json:"remote_type"`
	RemoteHost    string        `json:"remote_host"`
	RemotePort    int           `json:"remote_port"`
	Username      string        `json:"username"`
	Password      string        `json:"password,omitempty"`
	SSHPrivateKey string        `json:"ssh_private_key,omitempty"` // path to key or empty
	AuthMode      AuthMode      `json:"auth_mode"`                  // 认证模式：密码、匿名、私钥
	RemotePath    string        `json:"remote_path"`
	Module        string        `json:"module,omitempty"` // rsync daemon module
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
	State         TaskState     `json:"state"`
	LastError     string        `json:"last_error,omitempty"`
	Attempt       int           `json:"attempt"`
	// internal runtime fields (not persisted) could be omitted in JSON if desired
}

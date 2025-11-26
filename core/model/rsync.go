package model

type RsyncListTaskRequest struct {
	Page     int `form:"page" json:"page"`
	PageSize int `form:"page_size" json:"page_size"`
}

type RsyncListTaskResponse struct {
	Total int              `json:"total"`
	Tasks []*RsyncTaskInfo `json:"tasks"`
}

type RsyncTaskInfo struct {
	ID        string `json:"id"`
	Src       string `json:"src"`
	Dst       string `json:"dst"`
	CacheDir  string `json:"cache_dir"`
	Mode      string `json:"mode"`
	Status    string `json:"status"`
	Progress  int    `json:"progress"`
	Step      string `json:"step"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Error     string `json:"error"`
	LastLog   string `json:"last_log"`
}

type RsyncHost struct {
	ID       uint64 `json:"id"`
	Host     string `json:"host"`      // ip or hostname
	Port     int    `json:"port"`      // ssh port
	User     string `json:"user"`      // ssh user
	AuthMode string `json:"auth_mode"` // auth mode
	KeyPath  string `json:"key_path"`  // optional private key path on center
	Password string `json:"password"`  // optional (not recommended to store plaintext)
}

type RsyncCreateTaskRequest struct {
	Src     string    `form:"src" json:"src"`
	SrcHost RsyncHost `form:"src_host" json:"src_host"`
	Dst     string    `form:"dst" json:"dst"`
	DstHost RsyncHost `form:"dst_host" json:"dst_host"`
	Mode    string    `form:"mode" json:"mode"`
}

type RsyncCreateTask struct {
	SrcHostId int    `json:"src_host_id" validate:"required"`
	Src       string `json:"src" validate:"required"`
	DstHostId int    `json:"dst_host_id" validate:"required"`
	Dst       string `json:"dst" validate:"required"`
	Mode      string `json:"mode" validate:"required,oneof=copy incremental"`
}

type RsyncCreateTaskResponse struct {
	ID string `json:"id"`
}

type RsyncQueryTaskRequest struct {
	ID string `form:"id" json:"id"`
}

type RsyncCancelTaskRequest struct {
	ID string `form:"id" json:"id"`
}

type RsyncDeleteTaskRequest struct {
	ID string `form:"id" json:"id"`
}

type RsyncRetryTaskRequest struct {
	ID string `form:"id" json:"id"`
}

type RsyncClientCreateTaskRequest struct {
	Name          string `form:"name" json:"name" validate:"required"`
	Direction     string `form:"direction" json:"direction" validate:"required"`
	LocalPath     string `form:"local_path" json:"local_path" validate:"required"`
	RemoteType    string `form:"remote_type" json:"remote_type" validate:"required"`
	RemoteHost    string `form:"remote_host" json:"remote_host" validate:"required"`
	RemotePort    int    `form:"remote_port" json:"remote_port" validate:"required"`
	Username      string `form:"username" json:"username" validate:"required"`
	Password      string `form:"password" json:"password"`
	SSHPrivateKey string `form:"ssh_private_key" json:"ssh_private_key"`
	RemotePath    string `form:"remote_path" json:"remote_path" validate:"required"`
	Module        string `form:"module" json:"module"`
	Enqueue       bool   `form:"enqueue" json:"enqueue"` // whether to start immediately
}

type RsyncClientTask struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Direction     string `json:"direction"`
	LocalPath     string `json:"local_path"`
	RemoteType    string `json:"remote_type"`
	RemoteHost    string `json:"remote_host"`
	RemotePort    int    `json:"remote_port"`
	Username      string `json:"username"`
	Password      string `json:"password,omitempty"`
	SSHPrivateKey string `json:"ssh_private_key,omitempty"` // path to key or empty
	AuthMode      string `json:"auth_mode"`                 // 认证模式：密码、匿名、私钥
	RemotePath    string `json:"remote_path"`
	Module        string `json:"module,omitempty"` // rsync daemon module
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
	State         string `json:"state"`
	LastError     string `json:"last_error,omitempty"`
	Attempt       int    `json:"attempt"`
	// internal runtime fields (not persisted) could be omitted in JSON if desired
}

type RsyncClientListTaskResponse struct {
	Total int                `json:"total"`
	Tasks []*RsyncClientTask `json:"tasks"`
}

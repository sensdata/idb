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

type RsyncTestTaskRequest struct {
	ID string `form:"id" json:"id"`
}

type RsyncClientCreateTaskRequest struct {
	Name          string `json:"name" validate:"required"`
	Direction     string `json:"direction" validate:"required,oneof=remote_to_local local_to_remote"`
	LocalPath     string `json:"local_path" validate:"required"`
	RemoteType    string `json:"remote_type" validate:"required,oneof=ssh rsync"`
	RemoteHost    string `json:"remote_host" validate:"required"`
	RemotePort    int    `json:"remote_port" validate:"required"`
	Username      string `json:"username" validate:"required"`
	AuthMode      string `json:"auth_mode" validate:"required,oneof=password anonymous private_key"`
	Password      string `json:"password"`
	SSHPrivateKey string `json:"ssh_private_key"`
	RemotePath    string `json:"remote_path" validate:"required"`
	Module        string `json:"module"`
	Enqueue       bool   `json:"enqueue" validate:"required"` // whether to start immediately
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

type RsyncTaskLogListRequest struct {
	ID       string `form:"id" json:"id"`
	Page     int    `form:"page" json:"page"`
	PageSize int    `form:"page_size" json:"page_size"`
}

type RsyncTaskLogListResponse struct {
	Total int             `json:"total"`
	Logs  []*RsyncTaskLog `json:"logs"`
}

type RsyncTaskLog struct {
	ID   string `json:"id"`
	Path string `json:"path"`
}

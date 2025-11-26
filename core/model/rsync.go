package model

type RsyncListTaskRequest struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
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
	Src     string    `json:"src"`
	SrcHost RsyncHost `json:"src_host"`
	Dst     string    `json:"dst"`
	DstHost RsyncHost `json:"dst_host"`
	Mode    string    `json:"mode"`
}

type RsyncCreateTask struct {
	SrcHostId int    `json:"src_host_id" validate:"required"`
	Src       string `json:"src" validate:"required"`
	DstHostId int    `json:"dst_host_id" validate:"required"`
	Dst       string `json:"dst" validate:"required"`
	Mode      string `json:"mode" validate:"required,oneof=copy incremental"`
}

type RsyncClientCreateTaskRequest struct {
	Name          string `json:"name" validate:"required"`
	Direction     string `json:"direction" validate:"required"`
	LocalPath     string `json:"local_path" validate:"required"`
	RemoteType    string `json:"remote_type" validate:"required"`
	RemoteHost    string `json:"remote_host" validate:"required"`
	RemotePort    int    `json:"remote_port" validate:"required"`
	Username      string `json:"username" validate:"required"`
	Password      string `json:"password"`
	SSHPrivateKey string `json:"ssh_private_key"`
	RemotePath    string `json:"remote_path" validate:"required"`
	Module        string `json:"module"`
	Enqueue       bool   `json:"enqueue"` // whether to start immediately
}

type RsyncCreateTaskResponse struct {
	ID string `json:"id"`
}

type RsyncQueryTaskRequest struct {
	ID string `json:"id"`
}

type RsyncCancelTaskRequest struct {
	ID string `json:"id"`
}

type RsyncDeleteTaskRequest struct {
	ID string `json:"id"`
}

type RsyncRetryTaskRequest struct {
	ID string `json:"id"`
}

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
	Host     string `json:"host"`     // ip or hostname
	Port     int    `json:"port"`     // ssh port
	User     string `json:"user"`     // ssh user
	KeyPath  string `json:"key_path"` // optional private key path on center
	Password string `json:"password"` // optional (not recommended to store plaintext)
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

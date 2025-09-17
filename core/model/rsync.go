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

type RsyncCreateTaskRequest struct {
	Src  string `json:"src"`
	Dst  string `json:"dst"`
	Mode string `json:"mode"`
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

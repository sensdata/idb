package model

type ProcessSummary struct {
	PID         int32   `json:"pid"`
	PPID        int32   `json:"ppid"`
	Name        string  `json:"name"`
	CPUPercent  float64 `json:"cpu_percent"`
	MemPercent  float64 `json:"mem_percent"`
	MemRSS      uint64  `json:"mem_rss"`
	Swap        uint64  `json:"swap"`
	DiskRead    uint64  `json:"disk_read"`
	DiskWrite   uint64  `json:"disk_write"`
	NetSent     uint64  `json:"net_sent"`
	NetRecv     uint64  `json:"net_recv"`
	Connections int     `json:"connections"`
	User        string  `json:"user"`
	Threads     int32   `json:"threads"`
	CreateTime  int64   `json:"create_time"`
}

type ProcessDetail struct {
	Basic     ProcessBasicInfo  `json:"basic"`
	Memory    ProcessMemoryInfo `json:"memory"`
	Envs      []string          `json:"envs"`
	NetConns  []ProcessNetConn  `json:"net_conns"`
	OpenFiles []ProcessOpenFile `json:"open_files"`
}

type ProcessBasicInfo struct {
	Name        string `json:"name"`
	Status      string `json:"status"`
	PID         int32  `json:"pid"`
	PPID        int32  `json:"ppid"`
	Threads     int32  `json:"threads"`
	Connections int    `json:"connections"`
	DiskRead    uint64 `json:"disk_read"`
	DiskWrite   uint64 `json:"disk_write"`
	User        string `json:"user"`
	CreateTime  int64  `json:"create_time"`
	Cmdline     string `json:"cmdline"`
	Exe         string `json:"exe"`
	Cwd         string `json:"cwd"`
}

type ProcessMemoryInfo struct {
	RSS    uint64 `json:"rss"`
	Swap   uint64 `json:"swap"`
	VMS    uint64 `json:"vms"`
	HWM    uint64 `json:"hwm"`
	Data   uint64 `json:"data"`
	Stack  uint64 `json:"stack"`
	Locked uint64 `json:"locked"`
}

type ProcessNetConn struct {
	Protocol   string `json:"protocol"`
	LocalAddr  string `json:"local_addr"`
	LocalPort  uint32 `json:"local_port"`
	RemoteAddr string `json:"remote_addr"`
	RemotePort uint32 `json:"remote_port"`
	Status     string `json:"status"`
}

type ProcessOpenFile struct {
	Path string `json:"path"`
}

type ProcessListRequest struct {
	PageInfo
	OrderBy string `form:"order_by" json:"order_by" validate:"required,oneof=pid, name, cpu, mem"`
	Order   string `form:"order" json:"order" validate:"required,oneof=asc, desc"`
	Name    string `form:"name" json:"name"`
	Pid     int32  `form:"pid" json:"pid"`
	User    string `form:"user" json:"user"`
}

type ProcessListResponse struct {
	Total int64             `json:"total"`
	Items []*ProcessSummary `json:"items"`
}

type ProcessRequest struct {
	PID int32 `json:"pid"`
}

package model

// 系统信息
type SystemInfo struct {
	HostName            string `json:"host_name"`            //主机名称
	FQDN                string `json:"fqdn"`                 //主机完整域名
	Distribution        string `json:"distribution"`         //发行版
	DistributionVersion string `json:"distribution_version"` //发行版版本
	Version             string `json:"version"`              //兼容字段：系统版本
	Kernel              string `json:"kernel"`               //内核版本
	Platform            string `json:"platform"`             //兼容字段：发行版
	Arch                string `json:"arch"`                 //系统架构 x86_64, arm64, etc.
	OS                  string `json:"os"`                   //操作系统家族 linux/windows/darwin
	Virtual             string `json:"virtual"`              //虚拟化平台
	Vertual             string `json:"vertual"`              //兼容字段：虚拟化平台(拼写历史遗留)
	Uptime              int64  `json:"uptime"`               //运行时长(秒)
	MachineID           string `json:"machine_id,omitempty"` //机器标识
}

// 设置时间请求
type SetTimeReq struct {
	Timestamp int64 `json:"timestamp"` //时间戳
}

// 设置时区请求
type SetTimezoneReq struct {
	Timezone string `json:"timezone"` //时区
}

type AutoClearMemCacheReq struct {
	Interval int `json:"interval"` //时间间隔
}

type AutoClearMemCacheConf struct {
	Interval int `json:"interval"` //时间间隔
}

type CreateSwapReq struct {
	Size int `json:"size"` //大小，单位MB
}

type UpdateDnsSettingsReq struct {
	Servers []string `json:"servers"` //DNS服务器
	Timeout int      `json:"timeout"` //超时时间
	Retry   int      `json:"retry"`   //重试次数
}

type UpdateHostNameReq struct {
	HostName string `json:"host_name" validate:"required"` //主机名称
}

type UpdateSystemSettingsReq struct {
	MaxWatchFiles       int    `json:"max_watch_files" validate:"required,min=8192"`                         // inotify 监控的最大文件数
	MaxWatchInstances   int    `json:"max_watch_instances" validate:"required,min=128"`                      // inotify 实例数上限
	MaxQueuedEvents     int    `json:"max_queued_events" validate:"required,min=16384"`                      // inotify 队列长度上限
	MaxOpenFiles        int    `json:"max_open_files" validate:"required,min=1024"`                          // 系统最大打开文件数
	FileMax             int    `json:"file_max" validate:"required,min=65535"`                               // 系统文件句柄上限
	Swappiness          int    `json:"swappiness" validate:"min=0"`                                          // 交换分区倾向
	MaxMapCount         int    `json:"max_map_count" validate:"required,min=65530"`                          // 进程最大虚拟内存映射数量
	Somaxconn           int    `json:"somaxconn" validate:"required,min=128"`                                // socket 监听队列上限
	TcpMaxSynBacklog    int    `json:"tcp_max_syn_backlog" validate:"required,min=128"`                      // TCP SYN 队列上限
	PidMax              int    `json:"pid_max" validate:"required,min=32768"`                                // PID 上限
	OvercommitMemory    int    `json:"overcommit_memory" validate:"min=0,max=2"`                             // 内存 overcommit 策略
	OvercommitRatio     int    `json:"overcommit_ratio" validate:"min=0,max=100"`                            // overcommit 比例
	TransparentHugePage string `json:"transparent_huge_page" validate:"required,oneof=always madvise never"` // THP 策略
}

type SystemSettings struct {
	MaxWatchFiles       int    `json:"max_watch_files"`       // inotify 监控的最大文件数
	MaxWatchInstances   int    `json:"max_watch_instances"`   // inotify 实例数上限
	MaxQueuedEvents     int    `json:"max_queued_events"`     // inotify 队列长度上限
	MaxOpenFiles        int    `json:"max_open_files"`        // 系统最大打开文件数
	FileMax             int    `json:"file_max"`              // 系统文件句柄上限
	Swappiness          int    `json:"swappiness"`            // 交换分区倾向
	MaxMapCount         int    `json:"max_map_count"`         // 进程最大虚拟内存映射数量
	Somaxconn           int    `json:"somaxconn"`             // socket 监听队列上限
	TcpMaxSynBacklog    int    `json:"tcp_max_syn_backlog"`   // TCP SYN 队列上限
	PidMax              int    `json:"pid_max"`               // PID 上限
	OvercommitMemory    int    `json:"overcommit_memory"`     // 内存 overcommit 策略
	OvercommitRatio     int    `json:"overcommit_ratio"`      // overcommit 比例
	TransparentHugePage string `json:"transparent_huge_page"` // THP 策略
}

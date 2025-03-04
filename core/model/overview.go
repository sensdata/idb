package model

// 概览
type Overview struct {
	ServerTime     string      `json:"server_time"`      //服务器时间
	ServerTimeZone string      `json:"server_time_zone"` //服务器时区
	BootTime       string      `json:"boot_time"`        //启动时间
	RunTime        int64       `json:"run_time"`         //运行时间
	IdleTime       int64       `json:"idle_time"`        //空闲时间
	IdleRate       float64     `json:"idle_rate"`        //空闲率
	CpuUsage       string      `json:"cpu_usage"`        //CPU使用率
	CurrentLoad    LoadState   `json:"current_load"`     //当前负载
	MemoryUsage    MemoryState `json:"memory_usage"`     //内存使用
	SwapUsage      SwapState   `json:"swap_usage"`       //交换分区
	Storage        []Partition `json:"storage"`          //存储空间
}

// 负载
type LoadState struct {
	ProcessCount1  string `json:"process_count1"`  //1分钟进程数
	ProcessCount5  string `json:"process_count5"`  //5分钟进程数
	ProcessCount15 string `json:"process_count15"` //15分钟进程数
}

// 内存信息
type MemoryState struct {
	Physical string  `json:"physical"`  //物理内存
	Kernel   string  `json:"kernel"`    //内核占用
	Total    string  `json:"total"`     //总可用 = 物理内存 - 内核占用
	Free     string  `json:"free"`      //剩余可用
	FreeRate float64 `json:"free_rate"` //剩余率
	Used     string  `json:"used"`      //已使用 = 进程占用 + 缓冲区 + 缓存区
	UsedRate float64 `json:"used_rate"` //使用率
	Buffered string  `json:"buffered"`  //缓冲区
	Cached   string  `json:"cached"`    //缓存区
	RealUsed string  `json:"real_used"` //进程占用
}

// 交换分区信息
type SwapState struct {
	Total    string  `json:"total"`     //总大小
	Free     string  `json:"free"`      //剩余可用
	FreeRate float64 `json:"free_rate"` //剩余率
	Used     string  `json:"used"`      //已使用
	UsedRate float64 `json:"used_rate"` //使用率
}

// 分区
type Partition struct {
	Name     string  `json:"name"`      //名称
	Total    string  `json:"total"`     //总大小
	Free     string  `json:"free"`      //剩余可用
	Used     string  `json:"used"`      //已使用
	UsedRate float64 `json:"used_rate"` //使用率
}

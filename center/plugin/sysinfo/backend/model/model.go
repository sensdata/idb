package model

import "time"

type Summary struct {
	ServerTime     time.Time   `json:"serverTime"`     //服务器时间
	ServerTimeZone string      `json:"serverTimeZone"` //服务器时区
	BootTime       time.Time   `json:"bootTime"`       //启动时间
	RunTime        string      `json:"runTime"`        //运行时间
	IdleTime       string      `json:"idleTime"`       //空闲时间
	CpuUsage       int         `json:"cpuUsage"`       //CPU使用率
	CurrentLoad    LoadState   `json:"currentLoad"`    //当前负载
	MemoryUsage    MemoryState `json:"memoryUsage"`    //内存使用
	SwapUsage      SwapState   `json:"swapUsage"`      //交换分区
	Storage        []Partition `json:"storage"`        //存储空间
}

// 负载
type LoadState struct {
	ProcessCount1  float64 `json:"processCount1"`  //1分钟进程数
	ProcessCount5  float64 `json:"processCount5"`  //5分钟进程数
	ProcessCount15 float64 `json:"processCount15"` //15分钟进程数
}

// 内存信息
type MemoryState struct {
	Physical string  `json:"physical"` //物理内存
	Kernel   string  `json:"kernel"`   //内核占用
	Total    string  `json:"total"`    //总可用 = 物理内存 - 内核占用
	Free     string  `json:"free"`     //剩余可用
	FreeRate float64 `json:"freeRate"` //剩余率
	Used     string  `json:"used"`     //已使用 = 进程占用 + 缓冲区 + 缓存区
	UsedRate float64 `json:"usedRate"` //使用率
	Buffered string  `json:"buffered"` //缓冲区
	Cached   string  `json:"cached"`   //缓存区
	RealUsed string  `json:"realUsed"` //进程占用
}

// 交换分区信息
type SwapState struct {
	Total    string  `json:"total"`    //总大小
	Free     string  `json:"free"`     //剩余可用
	FreeRate float64 `json:"freeRate"` //剩余率
	Used     string  `json:"used"`     //已使用
	UsedRate float64 `json:"usedRate"` //使用率
}

// 分区
type Partition struct {
	Name     string  `json:"name"`     //名称
	Total    string  `json:"total"`    //总大小
	Free     string  `json:"free"`     //剩余可用
	Used     string  `json:"used"`     //已使用
	UsedRate float64 `json:"usedRate"` //使用率
}

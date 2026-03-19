package model

type Heartbeat struct {
	Command            string  `json:"command"`
	Cpu                float64 `json:"cpu"`
	Memory             float64 `json:"mem"`
	MemTotal           string  `json:"mem_total"` //总可用 = 物理内存 - 内核占用
	MemUsed            string  `json:"mem_used"`  //已使用 = 进程占用 + 缓冲区 + 缓存区
	Disk               float64 `json:"disk"`
	Rx                 float64 `json:"rx"` //接收实时速率
	Tx                 float64 `json:"tx"` //发送实时速率
	ProcessRSS         uint64  `json:"process_rss"`
	HeapAlloc          uint64  `json:"heap_alloc"`
	HeapSys            uint64  `json:"heap_sys"`
	StackInuse         uint64  `json:"stack_inuse"`
	Goroutines         int     `json:"goroutines"`
	OpenFDs            int     `json:"open_fds"`
	ActiveSessions     int     `json:"active_sessions"`
	ActiveLogFollowers int     `json:"active_log_followers"`
}

func NewHeartbeat() *Heartbeat {
	return &Heartbeat{
		Command:            "Heartbeat",
		Cpu:                0,
		Memory:             0,
		MemTotal:           "",
		MemUsed:            "",
		Disk:               0,
		Rx:                 0,
		Tx:                 0,
		ProcessRSS:         0,
		HeapAlloc:          0,
		HeapSys:            0,
		StackInuse:         0,
		Goroutines:         0,
		OpenFDs:            0,
		ActiveSessions:     0,
		ActiveLogFollowers: 0,
	}
}

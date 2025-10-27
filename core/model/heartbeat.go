package model

type Heartbeat struct {
	Command      string  `json:"command"`
	Cpu          float64 `json:"cpu"`
	Memory       float64 `json:"mem"`
	MemTotal     string  `json:"mem_total"` //总可用 = 物理内存 - 内核占用
	MemUsed      string  `json:"mem_used"`  //已使用 = 进程占用 + 缓冲区 + 缓存区
	Disk         float64 `json:"disk"`
	Rx           float64 `json:"rx"`             //接收实时速率
	Tx           float64 `json:"tx"`             //发送实时速率
	LastVerifyAt int64   `json:"last_verify_at"` // 上次验证时间
	Activated    bool    `json:"activated"`      // 是否已激活 !license.IssuedAt.IsZero()
}

func NewHeartbeat() *Heartbeat {
	return &Heartbeat{
		Command:      "Heartbeat",
		Cpu:          0,
		Memory:       0,
		MemTotal:     "",
		MemUsed:      "",
		Disk:         0,
		Rx:           0,
		Tx:           0,
		LastVerifyAt: 0,
		Activated:    false,
	}
}

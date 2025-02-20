package types

import "time"

// Metrics 系统监控指标
type Metrics struct {
	TaskCount      int64     `json:"task_count"`      // 总任务数
	ActiveTasks    int64     `json:"active_tasks"`    // 活跃任务数
	CompletedTasks int64     `json:"completed_tasks"` // 已完成任务数
	FailedTasks    int64     `json:"failed_tasks"`    // 失败任务数
	TotalLogSize   int64     `json:"total_log_size"`  // 总日志大小
	ReadOps        int64     `json:"read_ops"`        // 读操作次数
	WriteOps       int64     `json:"write_ops"`       // 写操作次数
	CleanupCount   int64     `json:"cleanup_count"`   // 清理操作次数
	ErrorCount     int64     `json:"error_count"`     // 错误次数
	LastError      string    `json:"last_error"`      // 最后一次错误信息
	LastErrorTime  time.Time `json:"last_error_time"` // 最后一次错误时间
	ActiveReaders  int64     `json:"active_readers"`  // 当前活跃读取器数量
	ActiveWriters  int64     `json:"active_writers"`  // 当前活跃写入器数量
}

package types

import (
	"fmt"
	"time"
)

// TaskStatus 任务状态
type TaskStatus string

const (
	TaskStatusCreated  TaskStatus = "created"
	TaskStatusRunning  TaskStatus = "running"
	TaskStatusSuccess  TaskStatus = "success"
	TaskStatusFailed   TaskStatus = "failed"
	TaskStatusCanceled TaskStatus = "canceled"
)

// TaskType 任务类型
const (
	TaskTypeBuffer = "buffer" // 内存缓冲日志
	TaskTypeFile   = "file"   // 文件存储日志
	TaskTypeRemote = "remote" // 远程存储日志
)

// Task 任务信息
type Task struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"` // buffer/file/remote
	Status    TaskStatus             `json:"status"`
	Metadata  map[string]interface{} `json:"metadata"`
	LogPath   string                 `json:"log_path"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// LogLevel 日志级别
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

// LogEntry 日志条目
type LogEntry struct {
	Timestamp time.Time         `json:"timestamp"`
	Level     LogLevel          `json:"level"`
	Message   string            `json:"message"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// ValidateTaskStatus 验证任务状态是否合法
func ValidateTaskStatus(status TaskStatus) error {
	switch status {
	case TaskStatusCreated,
		TaskStatusRunning,
		TaskStatusSuccess,
		TaskStatusFailed,
		TaskStatusCanceled:
		return nil
	default:
		return fmt.Errorf("%w: %s", ErrInvalidStatus, status)
	}
}

// ValidateStatusTransition 验证任务状态转换是否合法
func ValidateStatusTransition(current, new TaskStatus) error {
	// 已完成或取消的任务不能更改状态
	if current == TaskStatusSuccess ||
		current == TaskStatusFailed ||
		current == TaskStatusCanceled {
		return fmt.Errorf("%w: cannot change from %s", ErrInvalidStatus, current)
	}

	// 新建任务只能转为运行状态
	//if current == TaskStatusCreated && new != TaskStatusRunning {
	//	return fmt.Errorf("%w: created task can only transition to running", ErrInvalidStatus)
	//}

	// 运行中的任务只能转为完成、失败或取消状态
	if current == TaskStatusRunning {
		switch new {
		case TaskStatusSuccess,
			TaskStatusFailed,
			TaskStatusCanceled:
			return nil
		default:
			return fmt.Errorf("%w: running task can only transition to success/failed/canceled", ErrInvalidStatus)
		}
	}

	return nil
}

func (s TaskStatus) IsFinalStatus() bool {
	return s == TaskStatusSuccess ||
		s == TaskStatusFailed ||
		s == TaskStatusCanceled
}

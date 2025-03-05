package config

import (
	"path/filepath"
	"time"
)

// Config 日志流配置
type Config struct {
	// 基础路径配置
	BasePath string `json:"base_path"`
	LogDir   string `json:"log_dir"`
	TaskDir  string `json:"task_dir"`

	// 日志文件配置
	MaxLogSize    int64         `json:"max_log_size"`
	LogRetention  time.Duration `json:"log_retention"`
	LogBufferSize int           `json:"log_buffer_size"`

	// 任务配置
	TaskRetention time.Duration `json:"task_retention"`
	CleanInterval time.Duration `json:"clean_interval"`
	AutoCreateDir bool          `json:"auto_create_dir"`

	// 读取配置
	ReadBufferSize  int           `json:"read_buffer_size"`
	FollowInterval  time.Duration `json:"follow_interval"`
	MaxFollowBuffer int           `json:"max_follow_buffer"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		// 基础路径配置
		BasePath: filepath.Join("/var/lib/idb/data/logstream"),
		LogDir:   "logs",
		TaskDir:  "tasks",

		// 日志文件配置
		MaxLogSize:    100 * 1024 * 1024,  // 100MB
		LogRetention:  7 * 24 * time.Hour, // 7天
		LogBufferSize: 4096,

		// 任务配置
		TaskRetention: 30 * 24 * time.Hour, // 30天
		CleanInterval: 1 * time.Hour,       // 每小时清理一次
		AutoCreateDir: true,

		// 读取配置
		ReadBufferSize:  4096,
		FollowInterval:  100 * time.Millisecond,
		MaxFollowBuffer: 1000,
	}
}

// GetLogPath 获取日志文件完整路径
func (c *Config) GetLogPath(taskID string) string {
	return filepath.Join(c.BasePath, c.LogDir, taskID+".log")
}

// GetTaskPath 获取任务文件完整路径
func (c *Config) GetTaskPath(taskID string) string {
	return filepath.Join(c.BasePath, c.TaskDir, taskID+".json")
}

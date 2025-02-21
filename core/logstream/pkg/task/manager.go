package task

import (
	"bytes"
	"time"

	"github.com/sensdata/idb/core/logstream/pkg/types"
)

// Manager 任务管理器接口
type Manager interface {
	// Create 创建新任务
	Create(taskType string, metadata map[string]interface{}) (string, error)

	// Get 获取任务信息
	Get(taskID string) (*types.Task, error)

	// GetBuffer 获取任务缓冲区
	GetBuffer(taskID string) (*bytes.Buffer, error)

	// Update 更新任务状态
	Update(taskID string, status types.TaskStatus) error

	// Delete 删除任务
	Delete(taskID string) error

	// List 获取任务列表
	List() ([]*types.Task, error)

	// Clean 清理过期任务
	Clean(before time.Time) error
}

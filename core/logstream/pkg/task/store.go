package task

import (
	"github.com/sensdata/idb/core/logstream/pkg/types"
)

// Store 定义任务存储接口
type Store interface {
	// Save 保存任务信息
	Save(task *types.Task) error

	// Get 获取任务信息
	Get(taskID string) (*types.Task, error)

	// List 获取任务列表
	List() ([]*types.Task, error)

	// Delete 删除任务信息
	Delete(taskID string) error

	// Update 更新任务信息
	Update(task *types.Task) error
}

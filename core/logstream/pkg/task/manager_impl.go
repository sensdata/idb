package task

import (
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sensdata/idb/core/logstream/internal/config"
	"github.com/sensdata/idb/core/logstream/pkg/types"
)

type FileTaskManager struct {
	mu       sync.RWMutex
	tasks    map[string]*types.Task
	store    Store
	basePath string
}

func NewFileTaskManager(cfg *config.Config) (Manager, error) {
	store, err := NewFileStore(filepath.Join(cfg.BasePath, cfg.TaskDir))
	if err != nil {
		return nil, fmt.Errorf("create store failed: %v", err)
	}

	tasks := make(map[string]*types.Task)
	taskList, err := store.List()
	if err != nil {
		return nil, fmt.Errorf("load tasks failed: %v", err)
	}

	for _, task := range taskList {
		tasks[task.ID] = task
	}

	return &FileTaskManager{
		tasks:    tasks,
		store:    store,
		basePath: filepath.Join(cfg.BasePath, cfg.LogDir),
	}, nil
}

func (m *FileTaskManager) Create(taskType string, metadata map[string]interface{}) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	taskID := uuid.New().String()
	now := time.Now()

	task := &types.Task{
		ID:        taskID,
		Type:      taskType,
		Status:    types.TaskStatusCreated,
		Metadata:  metadata,
		LogPath:   filepath.Join(m.basePath, fmt.Sprintf("%s.log", taskID)),
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := m.store.Save(task); err != nil {
		return "", fmt.Errorf("save task failed: %v", err)
	}

	m.tasks[taskID] = task
	return taskID, nil
}

func (m *FileTaskManager) Get(taskID string) (*types.Task, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	task, exists := m.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task %s not found", taskID)
	}
	return task, nil
}

// validateTaskStatus 验证任务状态转换是否合法
func validateTaskStatus(current, new types.TaskStatus) error {
	// 已完成或取消的任务不能更改状态
	if current == types.TaskStatusSuccess ||
		current == types.TaskStatusFailed ||
		current == types.TaskStatusCanceled {
		return fmt.Errorf("task is already in final status: %s", current)
	}

	// 运行中的任务只能更新为完成、失败或取消状态
	if current == types.TaskStatusRunning {
		if new != types.TaskStatusSuccess &&
			new != types.TaskStatusFailed &&
			new != types.TaskStatusCanceled {
			return fmt.Errorf("invalid status transition: %s -> %s", current, new)
		}
	}

	return nil
}

func (m *FileTaskManager) Update(taskID string, status types.TaskStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, exists := m.tasks[taskID]
	if !exists {
		return fmt.Errorf("task %s not found", taskID)
	}

	if err := validateTaskStatus(task.Status, status); err != nil {
		return err
	}

	task.Status = status
	task.UpdatedAt = time.Now()

	if err := m.store.Update(task); err != nil {
		return fmt.Errorf("update task failed: %v", err)
	}

	return nil
}

func (m *FileTaskManager) Clean(before time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var cleanedCount int
	var lastError error

	for taskID, task := range m.tasks {
		if task.UpdatedAt.Before(before) {
			if err := m.store.Delete(taskID); err != nil {
				lastError = err
				continue
			}
			delete(m.tasks, taskID)
			cleanedCount++
		}
	}

	if lastError != nil {
		return fmt.Errorf("clean partially completed: cleaned %d tasks, last error: %v", cleanedCount, lastError)
	}

	return nil
}

func (m *FileTaskManager) Delete(taskID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.tasks[taskID]; !exists {
		return fmt.Errorf("task %s not found", taskID)
	}

	if err := m.store.Delete(taskID); err != nil {
		return fmt.Errorf("delete task failed: %v", err)
	}

	delete(m.tasks, taskID)
	return nil
}

func (m *FileTaskManager) List() ([]*types.Task, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tasks := make([]*types.Task, 0, len(m.tasks))
	for _, task := range m.tasks {
		tasks = append(tasks, task)
	}
	return tasks, nil
}

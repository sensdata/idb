package task

import (
	"bytes"
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
	store    Store
	basePath string
	tasks    map[string]*types.Task
	buffers  map[string]*bytes.Buffer
	watchers map[string][]*TaskStatusWatcher // 添加 watchers 管理
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
		basePath: filepath.Join(cfg.BasePath, cfg.LogDir),
		store:    store,
		tasks:    tasks,
		buffers:  make(map[string]*bytes.Buffer),
		watchers: make(map[string][]*TaskStatusWatcher),
	}, nil
}

func (m *FileTaskManager) Create(taskType string, metadata map[string]interface{}) (*types.Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	taskID := uuid.New().String()
	now := time.Now().Local()

	task := &types.Task{
		ID:        taskID,
		Type:      taskType,
		Status:    types.TaskStatusCreated,
		Metadata:  metadata,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// 根据任务类型设置日志路径
	switch taskType {
	case types.TaskTypeBuffer:
		task.LogPath = ""
		m.buffers[taskID] = bytes.NewBuffer(nil)
	case types.TaskTypeFile:
		// 从 metadata 中获取 log_path
		logPath, ok := metadata["log_path"]
		if !ok {
			// 如果没有设置log_path，则在默认目录下创建
			task.LogPath = filepath.Join(m.basePath, fmt.Sprintf("%s.log", taskID))
		} else {
			if logPathStr, ok := logPath.(string); ok {
				task.LogPath = logPathStr
			}
		}

	case types.TaskTypeRemote:
		// 从 metadata 中获取 log_path
		logPath, ok := metadata["log_path"]
		if !ok {
			return nil, fmt.Errorf("log_path not found in metadata")
		} else {
			if logPathStr, ok := logPath.(string); ok {
				task.LogPath = logPathStr
			}
		}
	}

	if err := m.store.Save(task); err != nil {
		delete(m.buffers, taskID) // 如果保存失败，清理已创建的 buffer
		return nil, fmt.Errorf("save task failed: %v", err)
	}

	m.tasks[taskID] = task
	return task, nil
}

func (m *FileTaskManager) Get(taskID string) (*types.Task, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	task, exists := m.tasks[taskID]
	if !exists {
		return nil, types.ErrTaskNotFound
	}
	return task, nil
}

func (m *FileTaskManager) GetByLog(logPath string) (*types.Task, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, task := range m.tasks {
		if task.LogPath == logPath {
			return task, nil
		}
	}
	return nil, types.ErrTaskNotFound
}

func (m *FileTaskManager) GetBuffer(taskID string) (*bytes.Buffer, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, exists := m.tasks[taskID]
	if !exists {
		return nil, types.ErrTaskNotFound
	}

	if task.Type != types.TaskTypeBuffer {
		return nil, fmt.Errorf("task %s is not buffer type", taskID)
	}

	if buf, exists := m.buffers[taskID]; exists {
		return buf, nil
	}

	buf := bytes.NewBuffer(nil)
	m.buffers[taskID] = buf
	return buf, nil
}

func (m *FileTaskManager) GetWatcher(taskID string) (TaskWatcher, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, exists := m.tasks[taskID]
	if !exists {
		return nil, types.ErrTaskNotFound
	}

	watcher := NewTaskStatusWatcher(taskID, m)
	if m.watchers[taskID] == nil {
		m.watchers[taskID] = make([]*TaskStatusWatcher, 0)
	}
	m.watchers[taskID] = append(m.watchers[taskID], watcher)
	return watcher, nil
}

func (m *FileTaskManager) Update(taskID string, status types.TaskStatus) error {
	m.mu.Lock()
	task, exists := m.tasks[taskID]
	if !exists {
		m.mu.Unlock()
		return fmt.Errorf("task %s not found", taskID)
	}

	if err := types.ValidateStatusTransition(task.Status, status); err != nil {
		m.mu.Unlock()
		return err
	}

	oldStatus := task.Status
	task.Status = status
	task.UpdatedAt = time.Now()

	if err := m.store.Update(task); err != nil {
		m.mu.Unlock()
		return fmt.Errorf("update task failed: %v", err)
	}

	// 只在状态真正发生变化时才通知观察者
	var watchersCopy []*TaskStatusWatcher
	if oldStatus != status {
		if watchers, exists := m.watchers[taskID]; exists {
			watchersCopy = append([]*TaskStatusWatcher(nil), watchers...)
		}
	}
	m.mu.Unlock()

	// 在锁外通知观察者
	for _, w := range watchersCopy {
		w.NotifyStatusChange(status)
	}

	// 任务完成时清理观察者
	if status.IsFinalStatus() {
		m.cleanClosedWatchers(taskID)
	}

	return nil
}

// removeWatcher 内部方法，从管理器中移除观察者
func (m *FileTaskManager) removeWatcher(taskID string, watcher *TaskStatusWatcher) {
	m.mu.Lock()
	defer m.mu.Unlock()

	watchers := m.watchers[taskID]
	for i, w := range watchers {
		if w == watcher {
			m.watchers[taskID] = append(watchers[:i], watchers[i+1:]...)
			break
		}
	}

	// 清理空的观察者列表
	if len(m.watchers[taskID]) == 0 {
		delete(m.watchers, taskID)
	}
}

// cleanClosedWatchers 清理已关闭的观察者
func (m *FileTaskManager) cleanClosedWatchers(taskID string) {
	m.mu.Lock()
	var taskIDs []string
	if taskID != "" {
		taskIDs = []string{taskID}
	} else {
		// 清理任务的所有观察者
		taskIDs = make([]string, 0, len(m.watchers))
		for id := range m.watchers {
			taskIDs = append(taskIDs, id)
		}
	}
	m.mu.Unlock()

	for _, id := range taskIDs {
		m.mu.Lock()
		watchers, exists := m.watchers[id]
		if !exists {
			m.mu.Unlock()
			continue
		}

		activeWatchers := make([]*TaskStatusWatcher, 0, len(watchers))
		for _, w := range watchers {
			if !w.IsClosed() {
				activeWatchers = append(activeWatchers, w)
			}
		}

		if len(activeWatchers) == 0 {
			delete(m.watchers, id)
		} else {
			m.watchers[id] = activeWatchers
		}
		m.mu.Unlock()
	}
}

func (m *FileTaskManager) Clean(before time.Time) error {
	// 先收集需要清理的任务
	m.mu.RLock()
	toClean := make([]string, 0)
	for taskID, task := range m.tasks {
		if task.UpdatedAt.Before(before) {
			toClean = append(toClean, taskID)
		}
	}
	m.mu.RUnlock()

	// 逐个清理任务
	var cleanedCount int
	var lastError error
	for _, taskID := range toClean {
		m.mu.Lock()
		// 关闭观察者
		if watchers, exists := m.watchers[taskID]; exists {
			for _, w := range watchers {
				w.Close()
			}
			delete(m.watchers, taskID)
		}

		// 删除任务
		if err := m.store.Delete(taskID); err != nil {
			lastError = err
			m.mu.Unlock()
			continue
		}
		delete(m.tasks, taskID)
		delete(m.buffers, taskID)
		cleanedCount++
		m.mu.Unlock()
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

	// 关闭并清理所有观察者
	if watchers, exists := m.watchers[taskID]; exists {
		for _, w := range watchers {
			w.Close()
		}
		delete(m.watchers, taskID)
	}

	delete(m.tasks, taskID)
	delete(m.buffers, taskID)
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

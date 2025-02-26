package task

import (
	"fmt"
	"sync"

	"github.com/sensdata/idb/core/logstream/pkg/types"
)

// TaskWatcher 任务状态监控接口
type TaskWatcher interface {
	// GetStatus 获取当前任务状态
	GetStatus() (types.TaskStatus, error)

	// WatchStatus 持续监控任务状态变化
	// 返回一个只读通道，用于接收状态更新
	WatchStatus() (<-chan types.TaskStatus, error)

	// Close 关闭监控器
	Close() error
}

const (
	defaultStatusChanSize = 10 // 状态通道的默认缓冲区大小
)

// TaskStatusWatcher 任务状态监控实现
type TaskStatusWatcher struct {
	mu      sync.RWMutex
	taskID  string
	manager Manager // 修改：移除指针，直接使用接口
	subs    []chan types.TaskStatus
	closed  bool
}

func NewTaskStatusWatcher(taskID string, manager Manager) *TaskStatusWatcher {
	return &TaskStatusWatcher{
		taskID:  taskID,
		manager: manager, // 修改：直接使用接口
		subs:    make([]chan types.TaskStatus, 0),
	}
}

func (w *TaskStatusWatcher) GetStatus() (types.TaskStatus, error) {
	task, err := w.manager.Get(w.taskID) // 修改：直接调用接口方法
	if err != nil {
		return "", err
	}
	return task.Status, nil
}

func (w *TaskStatusWatcher) WatchStatus() (<-chan types.TaskStatus, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return nil, fmt.Errorf("watcher is closed")
	}

	ch := make(chan types.TaskStatus, defaultStatusChanSize) // 修改：使用配置的缓冲区大小
	w.subs = append(w.subs, ch)

	// 获取并发送初始状态
	if status, err := w.GetStatus(); err == nil {
		ch <- status
	}

	return ch, nil
}

func (w *TaskStatusWatcher) NotifyStatusChange(status types.TaskStatus) {
	w.mu.RLock() // 添加：读锁保护
	defer w.mu.RUnlock()

	if w.closed { // 添加：检查是否已关闭
		return
	}

	for _, ch := range w.subs {
		select {
		case ch <- status:
		default:
			// 如果通道已满，跳过并记录日志
			// TODO: 添加日志记录
		}
	}
}

// IsClosed 返回监控器是否已关闭
func (w *TaskStatusWatcher) IsClosed() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.closed
}

// 内部接口
type watcherRemover interface {
	removeWatcher(taskID string, watcher *TaskStatusWatcher)
}

func (w *TaskStatusWatcher) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return nil
	}

	w.closed = true
	for _, ch := range w.subs {
		close(ch)
	}
	w.subs = nil

	// 使用接口的方式调用内部方法
	if remover, ok := w.manager.(watcherRemover); ok {
		remover.removeWatcher(w.taskID, w)
	}

	return nil
}

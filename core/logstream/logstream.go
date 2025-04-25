package logstream

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/sensdata/idb/core/logstream/internal/config"
	"github.com/sensdata/idb/core/logstream/internal/utils"
	"github.com/sensdata/idb/core/logstream/pkg/factory"
	"github.com/sensdata/idb/core/logstream/pkg/reader"
	"github.com/sensdata/idb/core/logstream/pkg/task"
	"github.com/sensdata/idb/core/logstream/pkg/types"
	"github.com/sensdata/idb/core/logstream/pkg/writer"
)

type LogStream struct {
	config  *config.Config
	taskMgr task.Manager
	metrics *types.MetricsCollector

	mu      sync.Mutex
	writers map[string]writer.Writer
	readers map[string]reader.Reader
}

func New(cfg *config.Config) (*LogStream, error) {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	// 确保目录存在
	if cfg.AutoCreateDir {
		if err := os.MkdirAll(filepath.Join(cfg.BasePath, cfg.LogDir), 0755); err != nil {
			return nil, fmt.Errorf("create log directory failed: %v", err)
		}
		if err := os.MkdirAll(filepath.Join(cfg.BasePath, cfg.TaskDir), 0755); err != nil {
			return nil, fmt.Errorf("create task directory failed: %v", err)
		}
	}

	// 创建任务管理器
	taskMgr, err := task.NewFileTaskManager(cfg)
	if err != nil {
		return nil, fmt.Errorf("create task manager failed: %v", err)
	}

	ls := &LogStream{
		config:  cfg,
		taskMgr: taskMgr,
		metrics: types.NewMetricsCollector(),
		writers: make(map[string]writer.Writer),
		readers: make(map[string]reader.Reader),
	}

	// 启动清理任务
	go ls.cleanRoutine()

	return ls, nil
}

func (ls *LogStream) GetTask(taskID string) (*types.Task, error) {
	task, err := ls.taskMgr.Get(taskID)
	if err != nil {
		if err == types.ErrTaskNotFound {
			return nil, err
		}
		ls.metrics.IncrErrorCount()
		return nil, fmt.Errorf("get task failed: %w", err)
	}
	return task, nil
}

func (ls *LogStream) GetWriter(taskID string) (writer.Writer, error) {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	if w, exists := ls.writers[taskID]; exists {
		return w, nil
	}

	task, err := ls.GetTask(taskID)
	if err != nil {
		return nil, err
	}

	w, err := factory.NewWriter(task, ls.taskMgr, ls.config)
	if err != nil {
		ls.metrics.IncrErrorCount()
		return nil, fmt.Errorf("create writer failed: %w", err)
	}

	ls.writers[taskID] = w
	ls.metrics.IncrWriteOps()
	return w, nil
}

func (ls *LogStream) GetReader(taskID string) (reader.Reader, error) {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	if r, exists := ls.readers[taskID]; exists {
		return r, nil
	}

	task, err := ls.GetTask(taskID)
	if err != nil {
		return nil, err
	}

	r, err := factory.NewReader(task, ls.taskMgr, ls.config)
	if err != nil {
		ls.metrics.IncrErrorCount()
		return nil, fmt.Errorf("create reader failed: %w", err)
	}

	ls.readers[taskID] = r
	ls.metrics.IncrReadOps()
	return r, nil
}

func (ls *LogStream) GetExistingReader(taskID string) (reader.Reader, error) {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	r, exists := ls.readers[taskID]
	if !exists {
		return nil, fmt.Errorf("reader not found for task: %s", taskID)
	}
	return r, nil
}

func (ls *LogStream) GetTaskWatcher(taskID string) (task.TaskWatcher, error) {
	watcher, err := ls.taskMgr.GetWatcher(taskID)
	if err != nil {
		return nil, err
	}

	return watcher, nil
}

func (ls *LogStream) CreateTask(taskType string, metadata map[string]interface{}) (*types.Task, error) {
	// 验证任务类型
	if taskType == "" {
		return nil, fmt.Errorf("task type cannot be empty")
	}

	// 验证元数据
	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	task, err := ls.taskMgr.Create(taskType, metadata)
	if err != nil {
		if err == types.ErrTaskExists {
			return nil, err
		}
		ls.metrics.IncrErrorCount()
		return nil, fmt.Errorf("create task failed: %w", err)
	}

	ls.metrics.IncrTaskCount()
	ls.metrics.IncrActiveTasks()
	return task, nil
}

func (ls *LogStream) UpdateTaskStatus(taskID string, status types.TaskStatus) error {
	// 验证状态转换是否合法
	if err := types.ValidateTaskStatus(status); err != nil {
		return fmt.Errorf("invalid task status: %w", err)
	}

	task, err := ls.GetTask(taskID)
	if err != nil {
		return err
	}

	// 验证状态转换
	if err := types.ValidateStatusTransition(task.Status, status); err != nil {
		return fmt.Errorf("invalid status transition: %w", err)
	}

	if err := ls.taskMgr.Update(taskID, status); err != nil {
		ls.metrics.IncrErrorCount()
		return fmt.Errorf("update task failed: %w", err)
	}

	// 更新指标
	switch status {
	case types.TaskStatusSuccess:
		ls.metrics.DecrActiveTasks()
		ls.metrics.IncrCompletedTasks()
	case types.TaskStatusFailed:
		ls.metrics.DecrActiveTasks()
		ls.metrics.IncrFailedTasks()
	}

	return nil
}

func (ls *LogStream) DeleteTask(taskID string) error {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	// 获取任务信息并检查状态
	task, err := ls.GetTask(taskID)
	if err != nil {
		return err
	}

	// 只允许删除已完成或失败的任务
	if task.Status != types.TaskStatusSuccess &&
		task.Status != types.TaskStatusFailed &&
		task.Status != types.TaskStatusCanceled {
		return fmt.Errorf("cannot delete task in %s status", task.Status)
	}

	// 关闭并删除相关的读写器
	if w, exists := ls.writers[taskID]; exists {
		w.Close()
		delete(ls.writers, taskID)
	}
	if r, exists := ls.readers[taskID]; exists {
		r.Close()
		delete(ls.readers, taskID)
	}

	// 删除任务
	if err := ls.taskMgr.Delete(taskID); err != nil {
		ls.metrics.IncrErrorCount()
		return fmt.Errorf("delete task failed: %w", err)
	}

	ls.metrics.DecrTaskCount()
	return nil
}

// GetConfig 获取当前配置
func (ls *LogStream) GetConfig() *config.Config {
	return ls.config
}

// GetLogPath 获取任务日志文件路径
func (ls *LogStream) GetLogPath(taskID string) string {
	return ls.config.GetLogPath(taskID)
}

// GetLogSize 获取任务日志文件大小
func (ls *LogStream) GetLogSize(taskID string) (int64, error) {
	logPath := ls.GetLogPath(taskID)
	return utils.GetFileSize(logPath)
}

// IsTaskExpired 检查任务是否过期
func (ls *LogStream) IsTaskExpired(taskID string) (bool, error) {
	task, err := ls.GetTask(taskID)
	if err != nil {
		return false, err
	}

	expiry := time.Now().Add(-ls.config.TaskRetention)
	return task.UpdatedAt.Before(expiry), nil
}

// GetTaskCount 获取当前任务数量
func (ls *LogStream) GetTaskCount() (int, error) {
	tasks, err := ls.taskMgr.List()
	if err != nil {
		return 0, err
	}
	return len(tasks), nil
}

// GetActiveReaders 获取当前活跃的读取器数量
func (ls *LogStream) GetActiveReaders() int {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	return len(ls.readers)
}

// GetActiveWriters 获取当前活跃的写入器数量
func (ls *LogStream) GetActiveWriters() int {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	return len(ls.writers)
}

func (ls *LogStream) Close() error {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	var errs []error

	// 关闭所有写入器
	for taskID, w := range ls.writers {
		if err := w.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close writer %s failed: %w", taskID, err))
			ls.metrics.IncrErrorCount()
			ls.metrics.RecordLastError(err)
		}
		delete(ls.writers, taskID)
	}

	// 关闭所有读取器
	for taskID, r := range ls.readers {
		if err := r.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close reader %s failed: %w", taskID, err))
			ls.metrics.IncrErrorCount()
			ls.metrics.RecordLastError(err)
		}
		delete(ls.readers, taskID)
	}

	// 如果有错误，合并所有错误信息
	if len(errs) > 0 {
		var errMsg string
		for i, err := range errs {
			if i > 0 {
				errMsg += "; "
			}
			errMsg += err.Error()
		}
		return fmt.Errorf("close errors: %s", errMsg)
	}

	return nil
}

func (ls *LogStream) CleanExpiredTasks() error {
	now := time.Now()

	// 清理过期任务
	taskExpiry := now.Add(-ls.config.TaskRetention)
	if err := ls.taskMgr.Clean(taskExpiry); err != nil {
		ls.metrics.IncrErrorCount()
		ls.metrics.RecordLastError(err)
		return fmt.Errorf("clean tasks failed: %w", err)
	}

	// 清理过期日志文件
	logExpiry := now.Add(-ls.config.LogRetention)
	logDir := filepath.Join(ls.config.BasePath, ls.config.LogDir)
	if err := utils.CleanOldFiles(logDir, logExpiry.Unix()); err != nil {
		ls.metrics.IncrErrorCount()
		ls.metrics.RecordLastError(err)
		return fmt.Errorf("clean logs failed: %w", err)
	}

	// 更新统计信息
	if err := ls.updateMetrics(); err != nil {
		ls.metrics.IncrErrorCount()
		ls.metrics.RecordLastError(err)
		return fmt.Errorf("update metrics failed: %w", err)
	}

	return nil
}

// GetMetrics 获取系统监控指标
func (ls *LogStream) GetMetrics() *types.Metrics {
	metrics := ls.metrics.GetMetrics()

	// 补充实时数据
	metrics.ActiveReaders = int64(ls.GetActiveReaders())
	metrics.ActiveWriters = int64(ls.GetActiveWriters())

	// 更新日志大小
	if tasks, err := ls.taskMgr.List(); err == nil {
		var totalSize int64
		for _, task := range tasks {
			if size, err := ls.GetLogSize(task.ID); err == nil {
				totalSize += size
			}
		}
		metrics.TotalLogSize = totalSize
	}

	return metrics
}

func (ls *LogStream) cleanRoutine() {
	ticker := time.NewTicker(ls.config.CleanInterval)
	defer ticker.Stop()

	for range ticker.C {
		if err := ls.CleanExpiredTasks(); err != nil {
			ls.metrics.IncrErrorCount()
			ls.metrics.RecordLastError(err)
			continue
		}
		ls.metrics.IncrCleanupCount()
	}
}

func (ls *LogStream) updateMetrics() error {
	tasks, err := ls.taskMgr.List()
	if err != nil {
		return err
	}

	var totalSize int64
	for _, task := range tasks {
		size, err := ls.GetLogSize(task.ID)
		if err != nil {
			continue
		}
		totalSize += size
	}
	ls.metrics.SetLogSize(totalSize)
	ls.metrics.SetTaskCount(int64(len(tasks)))
	return nil
}

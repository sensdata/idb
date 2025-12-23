package pkg

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sensdata/idb/agent/global"
)

// Manager controls tasks and execution
type Manager struct {
	storage        Storage
	mu             sync.RWMutex
	runtimeProcs   map[string]*ExecProcess
	queue          chan string // task IDs
	maxConcurrency int
	sem            chan struct{} // semaphore to limit concurrency
	stopCh         chan struct{}
	wg             sync.WaitGroup
}

// NewManager creates a manager; maxConcurrency default 1 if <=0
func NewManager(storage Storage, maxConcurrency int, queueSize int) *Manager {
	if maxConcurrency <= 0 {
		maxConcurrency = 1
	}
	m := &Manager{
		storage:        storage,
		runtimeProcs:   map[string]*ExecProcess{},
		queue:          make(chan string, queueSize),
		maxConcurrency: maxConcurrency,
		sem:            make(chan struct{}, maxConcurrency),
		stopCh:         make(chan struct{}),
	}

	// 状态修复已在FileJSONStorage初始化时完成，无需重复处理

	go m.dispatcher()
	return m
}

func (m *Manager) dispatcher() {
	for {
		select {
		case id := <-m.queue:
			// acquire semaphore
			m.sem <- struct{}{}
			m.wg.Add(1)
			go func(taskID string) {
				defer func() {
					<-m.sem
					m.wg.Done()
				}()
				if err := m.runTask(taskID); err != nil {
					// log
					global.LOG.Error("[rsyncmgr] runTask %s error: %v", taskID, err)
				}
			}(id)
		case <-m.stopCh:
			return
		}
	}
}

func (m *Manager) StopAll() {
	close(m.stopCh)
	m.wg.Wait()
}

// CreateTask create and persist task and enqueue it
func (m *Manager) CreateTask(t *RsyncTask, enqueue bool) (string, error) {
	if t == nil {
		return "", errors.New("nil task")
	}
	t.ID = uuid.New().String()
	now := time.Now()
	t.CreatedAt = now
	t.UpdatedAt = now
	t.State = StatePending
	t.Attempt = 0
	if err := m.storage.SaveTask(t); err != nil {
		global.LOG.Error("[rsyncmgr] failed to save task %s: %v", t.ID, err)
		return "", err
	}
	if enqueue {
		if err := m.EnqueueTask(t.ID); err != nil {
			global.LOG.Error("[rsyncmgr] failed to enqueue task %s: %v", t.ID, err)
			return "", err
		}
	}
	return t.ID, nil
}

func (m *Manager) EnqueueTask(id string) error {
	// validate exists
	if _, err := m.storage.GetTask(id); err != nil {
		global.LOG.Error("[rsyncmgr] failed to get task %s: %v", id, err)
		return err
	}
	select {
	case m.queue <- id:
		return nil
	default:
		// queue full, return error or block; we return error for now
		return errors.New("queue full")
	}
}

func (m *Manager) ListTasks(page, pageSize int) ([]*RsyncTask, error) {
	return m.storage.ListTasks(page, pageSize)
}

func (m *Manager) AllTasks() ([]*RsyncTask, error) {
	return m.storage.AllTasks()
}

func (m *Manager) GetTask(id string) (*RsyncTask, error) {
	return m.storage.GetTask(id)
}

func (m *Manager) DeleteTask(id string) error {
	// if running, stop first
	if proc := m.getProc(id); proc != nil {
		_ = proc.Stop()
	}
	return m.storage.DeleteTask(id)
}

func (m *Manager) StopTask(id string) error {
	proc := m.getProc(id)
	if proc == nil {
		// could be queued or not running
		// if queued, we can't remove easily from channel; instead update state
		t, err := m.storage.GetTask(id)
		if err != nil {
			global.LOG.Error("[rsyncmgr] failed to get task %s: %v", id, err)
			return err
		}
		if t.State == StateRunning {
			global.LOG.Error("[rsyncmgr] expected running but not found process for task %s", id)
			return errors.New("expected running but not found process")
		}
		t.State = StateStopped
		t.UpdatedAt = time.Now()
		return m.storage.SaveTask(t)
	}
	if err := proc.Stop(); err != nil {
		global.LOG.Error("[rsyncmgr] failed to stop rsync for task %s: %v", id, err)
		return err
	}
	// update state
	t, _ := m.storage.GetTask(id)
	t.State = StateStopped
	t.UpdatedAt = time.Now()
	return m.storage.SaveTask(t)
}

func (m *Manager) RetryTask(id string) error {
	t, err := m.storage.GetTask(id)
	if err != nil {
		return err
	}
	if t.State == StateRunning {
		global.LOG.Error("[rsyncmgr] cannot retry running task %s", id)
		return errors.New("cannot retry running task")
	}
	t.State = StatePending
	t.Attempt++
	t.UpdatedAt = time.Now()
	if err := m.storage.SaveTask(t); err != nil {
		global.LOG.Error("[rsyncmgr] failed to save task %s: %v", id, err)
		return err
	}
	return m.EnqueueTask(id)
}

// runTask executes one task lifecycle
func (m *Manager) runTask(id string) error {
	t, err := m.storage.GetTask(id)
	if err != nil {
		global.LOG.Error("[rsyncmgr] failed to get task %s: %v", id, err)
		return err
	}

	// 创建日志处理器
	logHandler := NewLogHandler(id)
	if logHandler == nil {
		global.LOG.Error("[rsyncmgr] failed to create log handler for task %s", id)
		return errors.New("failed to create log handler")
	}
	defer logHandler.Close()

	// 原子性操作：先启动进程，再更新状态和注册进程
	proc, err := StartRsync(t, logHandler)
	if err != nil {
		t.State = StateFailed
		t.LastError = err.Error()
		t.UpdatedAt = time.Now()
		if saveErr := m.storage.SaveTask(t); saveErr != nil {
			global.LOG.Error("[rsyncmgr] failed to save failed state for task %s: %v", id, saveErr)
			return fmt.Errorf("failed to start rsync: %v, and failed to save state: %v", err, saveErr)
		}

		// 记录错误信息到日志文件
		err = logHandler.AppendExecutionLog(fmt.Sprintf("Error: %v", err))
		if err != nil {
			global.LOG.Error("[rsyncmgr] failed to append execution log for task %s: %v", id, err)
			return err
		}

		global.LOG.Error("[rsyncmgr] failed to start rsync for task %s: %v", id, err)
		return err
	}

	// 原子性更新：在Manager锁保护下同时更新状态和注册进程
	m.mu.Lock()
	t.State = StateRunning
	t.UpdatedAt = time.Now()
	if saveErr := m.storage.SaveTask(t); saveErr != nil {
		m.mu.Unlock()
		proc.Stop() // 清理已启动的进程

		// 记录错误信息到日志文件
		err = logHandler.AppendExecutionLog(fmt.Sprintf("Error: %v", saveErr))
		if err != nil {
			global.LOG.Error("[rsyncmgr] failed to append execution log for task %s: %v", id, err)
			return fmt.Errorf("failed to save running state: %v", saveErr)
		}

		global.LOG.Error("[rsyncmgr] failed to save running state for task %s: %v", id, saveErr)
		return fmt.Errorf("failed to save running state: %v", saveErr)
	}
	m.runtimeProcs[id] = proc
	m.mu.Unlock()

	// wait for process finish
	err = proc.cmd.Wait()

	// 原子性清理：在Manager锁保护下同时更新状态和注销进程
	m.mu.Lock()
	defer m.mu.Unlock()

	// 重新获取最新状态的任务
	t, _ = m.storage.GetTask(id)
	if t != nil {
		if err != nil {
			t.State = StateFailed
			t.LastError = err.Error()
			global.LOG.Error("[rsyncmgr] rsync process failed for task %s: %v", id, err)
		} else {
			t.State = StateSucceeded
			t.LastError = ""
			global.LOG.Info("[rsyncmgr] rsync process completed successfully for task %s", id)
		}
		t.UpdatedAt = time.Now()
		if saveErr := m.storage.SaveTask(t); saveErr != nil {
			// 记录错误但不影响主流程
			global.LOG.Error("[rsyncmgr] failed to save final state for task %s: %v", id, saveErr)
		}
	}

	delete(m.runtimeProcs, id)
	return err
}

// TestSync 执行测试同步（dry-run），直接写入日志文件并返回日志路径
func (m *Manager) TestSync(id string) (string, error) {
	t, err := m.storage.GetTask(id)
	if err != nil {
		global.LOG.Error("[rsyncmgr] failed to get task %s for test sync: %v", id, err)
		return "", err
	}

	// 创建日志处理器
	logHandler := NewLogHandler(id)
	if logHandler == nil {
		global.LOG.Error("[rsyncmgr] failed to create log handler for test sync task %s", id)
		return "", errors.New("failed to create log handler")
	}
	defer logHandler.Close()

	// 执行测试同步，直接写入日志文件
	logPath, err := TestRsync(t, logHandler)
	if err != nil {
		// 记录错误信息到测试日志文件
		err = logHandler.AppendTestLog(fmt.Sprintf("Error: %v", err))
		if err != nil {
			global.LOG.Error("[rsyncmgr] failed to append test log for task %s: %v", id, err)
			return logPath, nil
		}
		global.LOG.Info("Test sync failed for task %s: %v", id, err)
		return logPath, nil
	}

	global.LOG.Info("[rsyncmgr] test sync completed for task %s, log saved to %s", id, logPath)
	return logPath, nil
}

func (m *Manager) GetTaskLogs(id string, page, pageSize int) (int, []string, error) {
	logHandler := NewLogHandler(id)
	if logHandler == nil {
		global.LOG.Error("[rsyncmgr] failed to create log handler for get logs task %s", id)
		return 0, nil, errors.New("failed to create log handler")
	}
	defer logHandler.Close()
	return logHandler.GetExecLogs(page, pageSize)
}

// helper proc map locks
//
//	func (m *Manager) setProc(id string, p *ExecProcess) {
//		m.mu.Lock()
//		defer m.mu.Unlock()
//		m.runtimeProcs[id] = p
//	}
//
//	func (m *Manager) deleteProc(id string) {
//		m.mu.Lock()
//		defer m.mu.Unlock()
//		delete(m.runtimeProcs, id)
//	}
func (m *Manager) getProc(id string) *ExecProcess {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.runtimeProcs[id]
}

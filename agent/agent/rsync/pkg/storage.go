package pkg

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/sensdata/idb/agent/global"
)

// Storage is an interface for persistence
type Storage interface {
	CreateTask(t *RsyncTask) error
	UpdateTask(t *RsyncTask) error
	GetTask(id string) (*RsyncTask, error)
	GetTaskByName(name string) (*RsyncTask, error)
	ListTasks(page, pageSize int) ([]*RsyncTask, error)
	AllTasks() ([]*RsyncTask, error)
	DeleteTask(id string) error
	Close() error
}

// InMemoryStorage simple implementation (good for testing)
type InMemoryStorage struct {
	mu    sync.RWMutex
	tasks map[string]*RsyncTask
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{tasks: map[string]*RsyncTask{}}
}

func (s *InMemoryStorage) CreateTask(t *RsyncTask) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 名称唯一性检查（仅新建任务）
	for id, existing := range s.tasks {
		if id == t.ID {
			return errors.New("task id already exists")
		}
		if existing.Name == t.Name {
			return errors.New("task name already exists")
		}
	}

	t.UpdatedAt = time.Now()
	s.tasks[t.ID] = t
	return nil
}

func (s *InMemoryStorage) UpdateTask(t *RsyncTask) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	t.UpdatedAt = time.Now()
	s.tasks[t.ID] = t
	return nil
}

func (s *InMemoryStorage) GetTask(id string) (*RsyncTask, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tasks[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return t, nil
}

func (s *InMemoryStorage) GetTaskByName(name string) (*RsyncTask, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, t := range s.tasks {
		if t.Name == name {
			return t, nil
		}
	}
	return nil, errors.New("not found")
}

func (s *InMemoryStorage) ListTasks(page, pageSize int) ([]*RsyncTask, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*RsyncTask, 0, len(s.tasks))
	for _, v := range s.tasks {
		out = append(out, v)
	}
	// 分页
	start := (page - 1) * pageSize
	end := start + pageSize
	if start >= len(out) {
		return []*RsyncTask{}, nil
	}
	if end > len(out) {
		end = len(out)
	}
	return out[start:end], nil
}

func (s *InMemoryStorage) AllTasks() ([]*RsyncTask, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*RsyncTask, 0, len(s.tasks))
	for _, v := range s.tasks {
		out = append(out, v)
	}
	return out, nil
}

func (s *InMemoryStorage) DeleteTask(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tasks, id)
	return nil
}

func (s *InMemoryStorage) Close() error { return nil }

// FileJSONStorage simple persistence using JSON file (safe for small number of tasks)
type FileJSONStorage struct {
	file string
	mu   sync.RWMutex
	// cache kept in memory for performance
	cache map[string]*RsyncTask
}

func NewFileJSONStorage(path string) (*FileJSONStorage, error) {
	// 1. 创建目录（如果不存在）
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage dir: %w", err)
	}

	s := &FileJSONStorage{
		file:  path,
		cache: map[string]*RsyncTask{},
	}

	// 2. 文件是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// 文件不存在，创建空文件并写入 {}
		if err := os.WriteFile(path, []byte("{}"), 0640); err != nil {
			return nil, fmt.Errorf("failed to create storage file: %w", err)
		}
		return s, nil
	}

	// 3. 文件存在 → 尝试读取
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read storage file: %w", err)
	}

	// 4. 尝试解析 JSON
	if len(b) > 0 {
		if err := json.Unmarshal(b, &s.cache); err != nil {
			// 文件损坏情况：重置为空结构
			global.LOG.Warn("storage file corrupted, resetting: %v", err)
			s.cache = map[string]*RsyncTask{}
			// 覆盖修复
			if err := os.WriteFile(path, []byte("{}"), 0640); err != nil {
				return nil, fmt.Errorf("failed to fix corrupted storage file: %w", err)
			}
		} else {
			// 文件解析成功，进行状态修复：将运行中的任务标记为失败
			for _, task := range s.cache {
				if task.State == StateRunning {
					task.State = StateFailed
					task.LastError = "agent restarted while task running"
					task.UpdatedAt = time.Now()
				}
			}
			// 保存修复后的状态
			if err := s.persistLocked(); err != nil {
				global.LOG.Warn("failed to save repaired task states: %v", err)
			}
		}
	}

	return s, nil
}

func (s *FileJSONStorage) persistLocked() error {
	b, err := json.MarshalIndent(s.cache, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.file, b, 0640)
}

func (s *FileJSONStorage) CreateTask(t *RsyncTask) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 名称唯一性检查（仅新建任务）
	for id, existing := range s.cache {
		if id == t.ID {
			return errors.New("task id already exists")
		}
		if existing.Name == t.Name {
			return errors.New("task name already exists")
		}
	}

	t.UpdatedAt = time.Now()
	s.cache[t.ID] = t
	return s.persistLocked()
}

func (s *FileJSONStorage) UpdateTask(t *RsyncTask) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	t.UpdatedAt = time.Now()
	s.cache[t.ID] = t
	return s.persistLocked()
}

func (s *FileJSONStorage) GetTask(id string) (*RsyncTask, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.cache[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return t, nil
}

func (s *FileJSONStorage) GetTaskByName(name string) (*RsyncTask, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, t := range s.cache {
		if t.Name == name {
			return t, nil
		}
	}
	return nil, errors.New("not found")
}

func (s *FileJSONStorage) ListTasks(page, pageSize int) ([]*RsyncTask, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*RsyncTask, 0, len(s.cache))
	for _, v := range s.cache {
		out = append(out, v)
	}
	// 分页
	start := (page - 1) * pageSize
	end := start + pageSize
	if start >= len(out) {
		return []*RsyncTask{}, nil
	}
	if end > len(out) {
		end = len(out)
	}
	return out[start:end], nil
}

func (s *FileJSONStorage) AllTasks() ([]*RsyncTask, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*RsyncTask, 0, len(s.cache))
	for _, v := range s.cache {
		out = append(out, v)
	}
	return out, nil
}

func (s *FileJSONStorage) DeleteTask(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.cache, id)
	return s.persistLocked()
}

func (s *FileJSONStorage) Close() error { return nil }

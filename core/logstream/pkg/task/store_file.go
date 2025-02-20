package task

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/sensdata/idb/core/logstream/pkg/types"
)

type FileStore struct {
	mu       sync.RWMutex
	basePath string
}

func NewFileStore(basePath string) (Store, error) {
	// 确保存储目录存在
	storePath := filepath.Join(basePath, "tasks")
	if err := os.MkdirAll(storePath, 0755); err != nil {
		return nil, fmt.Errorf("create store directory failed: %v", err)
	}

	return &FileStore{
		basePath: basePath,
	}, nil
}

func (s *FileStore) Save(task *types.Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("marshal task failed: %v", err)
	}

	filename := filepath.Join(s.basePath, "tasks", task.ID+".json")
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("write task file failed: %v", err)
	}

	return nil
}

func (s *FileStore) Get(taskID string) (*types.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	filename := filepath.Join(s.basePath, "tasks", taskID+".json")
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("task %s not found", taskID)
		}
		return nil, fmt.Errorf("read task file failed: %v", err)
	}

	var task types.Task
	if err := json.Unmarshal(data, &task); err != nil {
		return nil, fmt.Errorf("unmarshal task failed: %v", err)
	}

	return &task, nil
}

func (s *FileStore) List() ([]*types.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	dir := filepath.Join(s.basePath, "tasks")
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read directory failed: %v", err)
	}

	var tasks []*types.Task
	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		taskID := file.Name()[:len(file.Name())-5] // 移除 .json 后缀
		task, err := s.Get(taskID)
		if err != nil {
			continue
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (s *FileStore) Delete(taskID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	filename := filepath.Join(s.basePath, "tasks", taskID+".json")
	if err := os.Remove(filename); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("task %s not found", taskID)
		}
		return fmt.Errorf("delete task file failed: %v", err)
	}

	return nil
}

func (s *FileStore) Update(task *types.Task) error {
	return s.Save(task) // 直接覆盖原文件
}

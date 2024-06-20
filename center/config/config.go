package config

import (
	"encoding/json"
	"os"
	"sync"
)

// Config定义
type CenterConfig struct {
	Port      int    `json:"port"`
	DbPath    string `json:"db_path"`
	LogPath   string `json:"log_path"`
	SecretKey string `json:"secret_key"`
}

// Manager定义
type Manager struct {
	mu     sync.RWMutex
	config *CenterConfig
}

// 创建一个manager
func NewManager(configPath string) (*Manager, error) {
	manager := &Manager{}

	err := manager.loadConfig(configPath)
	if err != nil {
		return nil, err
	}

	// go manager.watchConfig(configPath)
	return manager, nil
}

// 加载配置文件
func (m *Manager) loadConfig(configPath string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	file, err := os.Open(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	var config CenterConfig
	err = decoder.Decode(&config)
	if err != nil {
		return err
	}

	m.config = &config
	return nil
}

// 获取当前配置
func (m *Manager) GetConfig() *CenterConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

package config

import (
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"
)

// Config定义
type Config struct {
	Port       int    `json:"port"`
	SecretKey  string `json:"secret_key"`
	CenterIP   string `json:"center_ip"`
	CenterPort int    `json:"center_port"`
}

// Manager定义
type Manager struct {
	mu     sync.RWMutex
	config *Config
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
	var config Config
	err = decoder.Decode(&config)
	if err != nil {
		return err
	}

	m.config = &config
	return nil
}

// 获取当前配置
func (m *Manager) GetConfig() *Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

// 监控配置文件的变化并热更新
func (m *Manager) watchConfig(configPath string) {
	lastModTime := time.Now()

	for {
		time.Sleep(time.Second * 5)

		fileInfo, err := os.Stat(configPath)
		if err != nil {
			log.Printf("Failed to stat config file: %v", err)
			continue
		}

		modTime := fileInfo.ModTime()
		if modTime.After(lastModTime) {
			log.Println("Config file changed, reloading...")
			err := m.loadConfig(configPath)
			if err != nil {
				log.Printf("Failed to reload config file: %v", err)
			} else {
				lastModTime = modTime
			}
		}
	}
}

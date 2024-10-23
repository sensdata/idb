package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
)

// Config定义
type CenterConfig struct {
	Host      string `json:"host"`
	Port      int    `json:"port"`
	SecretKey string `json:"secret_key"`
}

// Manager定义
type Manager struct {
	mu         sync.RWMutex
	config     *CenterConfig
	configPath string
}

// 创建一个manager
func NewManager(path string) (*Manager, error) {
	manager := &Manager{configPath: path}
	err := manager.loadConfig()
	if err != nil {
		return nil, err
	}

	return manager, nil
}

// 加载配置文件
func (m *Manager) loadConfig() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	file, err := os.Open(m.configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	config := &CenterConfig{}

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "#") {
			continue // 跳过空行或注释
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid config line: %s", line)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "host":
			config.Host = value
		case "port":
			fmt.Sscanf(value, "%d", &config.Port)
		case "secret_key":
			config.SecretKey = value
		default:
			return fmt.Errorf("unknown config key: %s", key)
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	m.config = config
	return nil
}

// 保存配置文件
func (m *Manager) saveConfig() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	file, err := os.Create(m.configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(m.config)
}

// 获取当前配置
func (m *Manager) GetConfig() *CenterConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

// 检查是否是支持的项
func validateItem(item string) bool {
	switch item {
	case "":
		return true
	case "host":
		return true
	case "port":
		return true
	case "key":
		return true
	default:
		return false
	}
}

// 获取当前配置的shell打印
func (m *Manager) GetConfigString(item string) (string, error) {
	if !validateItem(item) {
		return "", fmt.Errorf("%s not supported", item)
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	var result strings.Builder

	if item == "" {
		result.WriteString(fmt.Sprintf("host=%s\n", m.config.Host))
		result.WriteString(fmt.Sprintf("port=%d\n", m.config.Port))
		result.WriteString(fmt.Sprintf("secret_key=%s\n", m.config.SecretKey))
	} else {
		switch item {
		case "host":
			result.WriteString(fmt.Sprintf("host=%s\n", m.config.Host))
		case "port":
			result.WriteString(fmt.Sprintf("port=%d\n", m.config.Port))
		case "secret_key":
			result.WriteString(fmt.Sprintf("secret_key=%s\n", m.config.SecretKey))
		}
	}

	return result.String(), nil
}

// 设置配置项
func (m *Manager) SetConfig(key, value string) error {
	if !validateItem(key) {
		return fmt.Errorf("%s not supported", key)
	}

	switch key {
	case "host":
		m.config.Host = value
	case "port":
		portValue, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		m.config.Port = portValue
	case "secret_key":
		m.config.SecretKey = value
	}

	// 保存到文件
	return m.saveConfig()
}

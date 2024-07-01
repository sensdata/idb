package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

// Config定义
type Config struct {
	Port      int    `json:"port"`
	SecretKey string `json:"secret_key"`
}

// Manager定义
type Manager struct {
	mu         sync.RWMutex
	config     *Config
	configPath string
}

// 创建一个manager
func NewManager(installDir string) (*Manager, error) {
	manager := &Manager{}
	err := manager.loadConfig(installDir)
	if err != nil {
		return nil, err
	}

	return manager, nil
}

// 加载配置文件
func (m *Manager) loadConfig(installDir string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 配置文件路径
	m.configPath = filepath.Join(installDir, "config.json")

	file, err := os.Open(m.configPath)
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
func (m *Manager) GetConfig() *Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

// 检查是否是支持的项
func validateItem(item string) bool {
	switch item {
	case "":
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
		return "", fmt.Errorf("%s not support", item)
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	var result strings.Builder

	v := reflect.ValueOf(m.config).Elem()
	t := reflect.TypeOf(m.config).Elem()

	if item == "" {
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			tag := t.Field(i).Tag.Get("json")
			result.WriteString(fmt.Sprintf("%s: %v\n", tag, field.Interface()))
		}
	} else {
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			tag := t.Field(i).Tag.Get("json")
			if strings.ToLower(tag) == strings.ToLower(item) {
				result.WriteString(fmt.Sprintf("%s: %v\n", tag, field.Interface()))
				break
			}
		}
	}

	return result.String(), nil
}

// 设置配置项
func (m *Manager) SetConfig(key, value string) error {
	if !validateItem(key) {
		return fmt.Errorf("%s not support", key)
	}

	v := reflect.ValueOf(m.config).Elem()
	field := v.FieldByNameFunc(func(name string) bool {
		return strings.EqualFold(name, key)
	})

	if !field.IsValid() {
		return fmt.Errorf("invalid configuration key: %s", key)
	}

	// 转换 value 为 field 对应的类型
	switch field.Kind() {
	case reflect.Int:
		intValue, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		field.SetInt(int64(intValue))
	case reflect.String:
		field.SetString(value)
	default:
		return fmt.Errorf("unsupported configuration key: %s", key)
	}

	// 保存到文件
	return m.saveConfig()
}

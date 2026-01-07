package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
)

// Config定义
type CenterConfig struct {
	Host   string `json:"host"`
	Port   int    `json:"port"`
	Latest string `json:"latest"`
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

	// 检查配置文件格式，如果是 JSON 格式则转换为 key=value 格式
	// 通过读取文件前几个字节来判断
	file, err := os.Open(path)
	if err == nil {
		peekBytes := make([]byte, 1)
		file.Read(peekBytes)
		file.Close()
		// 如果是 JSON 格式，转换为 key=value 格式
		if len(peekBytes) > 0 && peekBytes[0] == '{' {
			// 释放锁后保存为标准格式
			if err := manager.saveConfig(); err != nil {
				// 转换失败不影响使用，只记录警告
				// 注意：这里不能使用 global.LOG，因为可能还未初始化
				// 可以考虑返回错误或使用标准库 log
			}
		}
	}

	return manager, nil
}

// 加载配置文件
// 支持 key=value 格式和 JSON 格式（向后兼容）
func (m *Manager) loadConfig() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	file, err := os.Open(m.configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 读取文件的前几个字节来检测格式
	peekBytes := make([]byte, 1)
	_, err = file.Read(peekBytes)
	if err != nil && err != io.EOF {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// 重置文件指针到开头
	file.Seek(0, 0)

	config := &CenterConfig{}

	// 检测是否为 JSON 格式（以 { 开头）
	if len(peekBytes) > 0 && peekBytes[0] == '{' {
		// JSON 格式（向后兼容旧版本）
		decoder := json.NewDecoder(file)
		if err := decoder.Decode(config); err != nil {
			return fmt.Errorf("failed to parse JSON config: %w", err)
		}
		// 如果成功解析 JSON，立即保存为 key=value 格式以统一格式
		m.config = config
		// 注意：这里不能直接调用 saveConfig()，因为已经在锁内，会导致死锁
		// 所以先设置 config，然后在外部调用 saveConfig() 来转换格式
		return nil
	}

	// key=value 格式（标准格式）
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
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
			portValue, err := strconv.Atoi(value)
			if err != nil {
				return fmt.Errorf("invalid port value: %s, err: %w", value, err)
			}
			config.Port = portValue
		case "latest":
			config.Latest = value
		case "admin_pass":
			// 向后兼容：忽略 admin_pass，不解析到结构体中
			// 密码不应该保存在配置文件中，如果需要读取，使用 GetAdminPassFromConfig()
			continue
		default:
			return fmt.Errorf("unknown config key: %s", key)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to scan config file: %w", err)
	}

	m.config = config
	return nil
}

// 保存配置文件
// 使用 key=value 格式保存，与 loadConfig 保持一致
// 保存前会创建备份，如果保存失败会尝试恢复备份
func (m *Manager) saveConfig() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 创建备份文件路径
	backupPath := m.configPath + ".bak"

	// 如果配置文件存在，先创建备份
	if _, err := os.Stat(m.configPath); err == nil {
		// 读取原文件内容
		srcFile, err := os.Open(m.configPath)
		if err != nil {
			return fmt.Errorf("failed to open config file for backup: %w", err)
		}
		defer srcFile.Close()

		// 创建备份文件
		dstFile, err := os.Create(backupPath)
		if err != nil {
			return fmt.Errorf("failed to create backup file: %w", err)
		}
		defer dstFile.Close()

		// 复制文件内容
		_, err = io.Copy(dstFile, srcFile)
		if err != nil {
			dstFile.Close()
			os.Remove(backupPath) // 清理失败的备份
			return fmt.Errorf("failed to backup config file: %w", err)
		}
	}

	// 创建临时文件用于写入新配置
	tmpPath := m.configPath + ".tmp"
	file, err := os.Create(tmpPath)
	if err != nil {
		// 如果创建临时文件失败，尝试恢复备份
		if _, err2 := os.Stat(backupPath); err2 == nil {
			os.Rename(backupPath, m.configPath)
		}
		return fmt.Errorf("failed to create temp config file: %w", err)
	}

	// 写入配置内容（key=value 格式）
	writeErr := func() error {
		defer file.Close()

		// 写入 host
		if _, err := fmt.Fprintf(file, "host=%s\n", m.config.Host); err != nil {
			return err
		}
		// 写入 port
		if _, err := fmt.Fprintf(file, "port=%d\n", m.config.Port); err != nil {
			return err
		}
		// 写入 latest（如果非空）
		if m.config.Latest != "" {
			if _, err := fmt.Fprintf(file, "latest=%s\n", m.config.Latest); err != nil {
				return err
			}
		}
		// 注意：不写入 admin_pass，因为密码不应该保存在配置文件中
		// admin_pass 仅在首次初始化时从配置文件读取（向后兼容），初始化完成后会被清理

		return file.Sync() // 确保数据写入磁盘
	}()

	if writeErr != nil {
		file.Close()
		os.Remove(tmpPath) // 清理临时文件
		// 尝试恢复备份
		if _, err2 := os.Stat(backupPath); err2 == nil {
			os.Rename(backupPath, m.configPath)
		}
		return fmt.Errorf("failed to write config file: %w", writeErr)
	}

	// 原子性替换：将临时文件重命名为正式文件
	if err := os.Rename(tmpPath, m.configPath); err != nil {
		os.Remove(tmpPath) // 清理临时文件
		// 尝试恢复备份
		if _, err2 := os.Stat(backupPath); err2 == nil {
			os.Rename(backupPath, m.configPath)
		}
		return fmt.Errorf("failed to rename temp config file: %w", err)
	}

	// 保存成功后，清理备份文件
	if _, err := os.Stat(backupPath); err == nil {
		os.Remove(backupPath)
	}

	return nil
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
	case "latest":
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
	} else {
		switch item {
		case "host":
			result.WriteString(fmt.Sprintf("host=%s\n", m.config.Host))
		case "port":
			result.WriteString(fmt.Sprintf("port=%d\n", m.config.Port))
		case "latest":
			result.WriteString(fmt.Sprintf("latest=%s\n", m.config.Latest))
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
	case "latest":
		m.config.Latest = value
	}

	// 保存到文件
	return m.saveConfig()
}

// GetAdminPassFromEnv 从环境变量读取管理员密码
// 优先从环境变量 PASSWORD 读取，如果不存在则尝试 ADMIN_PASS
// 如果都不存在则返回空字符串
func GetAdminPassFromEnv() string {
	// 优先从 PASSWORD 环境变量读取（docker-compose.yaml 中设置）
	if pass := os.Getenv("PASSWORD"); pass != "" {
		return pass
	}
	// 如果 PASSWORD 不存在，尝试从 ADMIN_PASS 读取（兼容其他部署方式）
	if pass := os.Getenv("ADMIN_PASS"); pass != "" {
		return pass
	}
	return ""
}

// RemoveAdminPass 从配置文件中移除 admin_pass 字段
// 用于初始化完成后清理敏感信息
func (m *Manager) RemoveAdminPass() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 读取配置文件内容
	file, err := os.Open(m.configPath)
	if err != nil {
		// 文件不存在或无法打开，直接返回（可能已经被删除）
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string
	hasAdminPass := false

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		// 跳过 admin_pass 行
		if strings.HasPrefix(trimmed, "admin_pass=") {
			hasAdminPass = true
			continue
		}
		lines = append(lines, line)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to scan config file: %w", err)
	}

	// 如果没有找到 admin_pass，无需修改
	if !hasAdminPass {
		return nil
	}

	// 创建备份
	backupPath := m.configPath + ".bak"
	if _, err := os.Stat(m.configPath); err == nil {
		srcFile, err := os.Open(m.configPath)
		if err == nil {
			defer srcFile.Close()
			dstFile, err := os.Create(backupPath)
			if err == nil {
				io.Copy(dstFile, srcFile)
				dstFile.Close()
			}
		}
	}

	// 写入新内容（不包含 admin_pass）
	tmpPath := m.configPath + ".tmp"
	tmpFile, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	for _, line := range lines {
		fmt.Fprintln(tmpFile, line)
	}

	if err := tmpFile.Sync(); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("failed to sync temp file: %w", err)
	}
	tmpFile.Close()

	// 原子性替换
	if err := os.Rename(tmpPath, m.configPath); err != nil {
		os.Remove(tmpPath)
		// 尝试恢复备份
		if _, err2 := os.Stat(backupPath); err2 == nil {
			os.Rename(backupPath, m.configPath)
		}
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	// 清理备份
	if _, err := os.Stat(backupPath); err == nil {
		os.Remove(backupPath)
	}

	return nil
}

// GetAdminPassFromConfig 从配置文件中读取 admin_pass（向后兼容）
// 仅在首次初始化时使用，初始化完成后应该被清理
// 如果配置文件中不存在 admin_pass，返回空字符串
func (m *Manager) GetAdminPassFromConfig() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	file, err := os.Open(m.configPath)
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "admin_pass=") {
			parts := strings.SplitN(trimmed, "=", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}

	return ""
}

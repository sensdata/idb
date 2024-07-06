package utils

import (
	"fmt"
	"os"
	"strconv"
	"syscall"
)

// CreatePIDFile 创建 PID 文件
func CreatePIDFile(pidFilePath string) error {
	pid := os.Getpid()
	file, err := os.Create(pidFilePath)
	if err != nil {
		return fmt.Errorf("failed to create PID file: %v", err)
	}
	defer file.Close()
	_, err = file.WriteString(strconv.Itoa(pid))
	if err != nil {
		return fmt.Errorf("failed to write PID to file: %v", err)
	}
	return nil
}

// RemovePIDFile 删除 PID 文件
func RemovePIDFile(pidFilePath string) error {
	return os.Remove(pidFilePath)
}

// IsRunning 判断服务是否已经启动
func IsRunning(pidFilePath string) (bool, error) {
	data, err := os.ReadFile(pidFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to read PID file: %v", err)
	}

	pid, err := strconv.Atoi(string(data))
	if err != nil {
		return false, fmt.Errorf("invalid PID in file: %v", err)
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return false, fmt.Errorf("failed to find process: %v", err)
	}

	// 发送信号 0 来检查进程是否存在
	err = process.Signal(syscall.Signal(0))
	if err != nil {
		if err == syscall.ESRCH {
			return false, nil
		}
		return false, fmt.Errorf("failed to signal process: %v", err)
	}

	return true, nil
}

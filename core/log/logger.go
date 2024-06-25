package log

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var (
	logger *log.Logger
)

func Writer() *log.Logger {
	return log.New(os.Stdout, "\r\n", log.LstdFlags)
}

// InitLogger 初始化日志
func InitLogger(logfilePath string) error {
	// Ensure the directory exists
	dir := filepath.Dir(logfilePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			fmt.Printf("Failed to create log directory: %v", err)
		}
	}

	// 打开日志文件
	logFile, err := os.OpenFile(logfilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	// 创建日志实例
	logger = log.New(logFile, "[Agent] ", log.LstdFlags|log.Lshortfile)
	return nil
}

// Info 记录信息日志
func Info(format string, args ...interface{}) {
	if logger != nil {
		// 手动添加换行符
		format += "\n"
		logger.Printf("[INFO] "+format, args...)
		fmt.Printf("[INFO] "+format, args...)
	}
}

// Error 记录错误日志
func Error(format string, args ...interface{}) {
	if logger != nil {
		// 手动添加换行符
		format += "\n"
		logger.Printf("[ERROR] "+format, args...)
		fmt.Printf("[ERROR] "+format, args...)
	}
}

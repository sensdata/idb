package pkg

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/sensdata/idb/agent/global"
	"github.com/sensdata/idb/core/constant"
)

// LogHandler 负责管理rsync任务的日志文件
type LogHandler struct {
	taskID      string
	logDir      string
	mu          sync.Mutex
	execLogFile *os.File
	testLogFile *os.File
}

// NewLogHandler 创建新的日志处理器
func NewLogHandler(taskID string) *LogHandler {
	logDir := filepath.Join(constant.AgentDataDir, "rsync", "logs", taskID)

	// 确保日志目录存在
	if err := os.MkdirAll(logDir, 0755); err != nil {
		global.LOG.Error("[rsyncmgr] failed to create log directory %s: %v", logDir, err)
		return nil
	}

	return &LogHandler{
		taskID: taskID,
		logDir: logDir,
	}
}

// AppendExecutionLog 追加执行日志（用于正常任务执行过程）
func (lh *LogHandler) AppendExecutionLog(message string) error {
	lh.mu.Lock()
	defer lh.mu.Unlock()

	if lh.execLogFile == nil {
		logPath := lh.getExecLogPath()
		file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			global.LOG.Error("[rsyncmgr] failed to open execution log file %s: %v", logPath, err)
			return err
		}
		lh.execLogFile = file
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logEntry := fmt.Sprintf("[%s] %s\n", timestamp, message)

	_, err := lh.execLogFile.WriteString(logEntry)
	if err != nil {
		global.LOG.Error("[rsyncmgr] failed to write execution log for task %s: %v", lh.taskID, err)
		return err
	}

	return nil
}

// GetExecutionLogWriter 获取执行日志的写入器，用于直接将命令输出重定向到日志文件
func (lh *LogHandler) GetExecutionLogWriter() (*os.File, error) {
	lh.mu.Lock()
	defer lh.mu.Unlock()

	if lh.execLogFile == nil {
		logPath := lh.getExecLogPath()
		file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			global.LOG.Error("[rsyncmgr] failed to open execution log file %s: %v", logPath, err)
			return nil, err
		}
		lh.execLogFile = file
	}

	return lh.execLogFile, nil
}

// GetTestLogWriter 获取测试日志的写入器，用于直接将测试命令输出重定向到日志文件
func (lh *LogHandler) GetTestLogWriter() (*os.File, error) {
	lh.mu.Lock()
	defer lh.mu.Unlock()

	if lh.testLogFile == nil {
		logPath := lh.getTestLogPath()
		file, err := os.Create(logPath)
		if err != nil {
			global.LOG.Error("[rsyncmgr] failed to create test log file %s: %v", logPath, err)
			return nil, err
		}
		lh.testLogFile = file
	}

	return lh.testLogFile, nil
}

// AppendTestLog 追加测试日志（用于测试同步过程中的错误记录等）
func (lh *LogHandler) AppendTestLog(message string) error {
	lh.mu.Lock()
	defer lh.mu.Unlock()

	if lh.testLogFile == nil {
		logPath := lh.getTestLogPath()
		file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			global.LOG.Error("[rsyncmgr] failed to open test log file %s: %v", logPath, err)
			return err
		}
		lh.testLogFile = file
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logEntry := fmt.Sprintf("[%s] %s\n", timestamp, message)

	_, err := lh.testLogFile.WriteString(logEntry)
	if err != nil {
		global.LOG.Error("[rsyncmgr] failed to write test log for task %s: %v", lh.taskID, err)
		return err
	}

	return nil
}

// Close 关闭所有日志文件
func (lh *LogHandler) Close() error {
	lh.mu.Lock()
	defer lh.mu.Unlock()

	var err error

	if lh.execLogFile != nil {
		if closeErr := lh.execLogFile.Close(); closeErr != nil {
			err = closeErr
		}
		lh.execLogFile = nil
	}

	if lh.testLogFile != nil {
		if closeErr := lh.testLogFile.Close(); closeErr != nil {
			err = closeErr
		}
		lh.testLogFile = nil
	}

	return err
}

// GetExecLogPath 获取执行日志文件路径
func (lh *LogHandler) GetExecLogPath() string {
	return lh.getExecLogPath()
}

// GetTestLogPath 获取测试日志文件路径
func (lh *LogHandler) GetTestLogPath() string {
	return lh.getTestLogPath()
}

// getExecLogPath 获取执行日志文件路径
func (lh *LogHandler) getExecLogPath() string {
	timestamp := time.Now().Format("20060102-150405")
	return filepath.Join(lh.logDir, fmt.Sprintf("%s.log", timestamp))
}

// getTestLogPath 获取测试日志文件路径
func (lh *LogHandler) getTestLogPath() string {
	timestamp := time.Now().Format("20060102-150405")
	return filepath.Join(lh.logDir, fmt.Sprintf("test-%s.log", timestamp))
}

// GetExecLogs 枚举任务的所有执行日志文件路径
func (lh *LogHandler) GetExecLogs(page, pageSize int) (int, []string, error) {
	// 枚举该目录下，全部非 test- 开头的.log文件
	files, err := os.ReadDir(lh.logDir)
	if err != nil {
		global.LOG.Error("[rsyncmgr] failed to read log directory %s: %v", lh.logDir, err)
		return 0, nil, err
	}
	var execLogs []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if filepath.Ext(file.Name()) != ".log" {
			continue
		}
		if strings.HasPrefix(file.Name(), "test-") {
			continue
		}
		execLogs = append(execLogs, filepath.Join(lh.logDir, file.Name()))
	}

	// 分页
	start := (page - 1) * pageSize
	end := page * pageSize
	if start >= len(execLogs) {
		return 0, []string{}, nil
	}
	if end > len(execLogs) {
		end = len(execLogs)
	}
	return len(execLogs), execLogs[start:end], nil
}

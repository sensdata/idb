package log

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Log struct {
	Logger *zap.Logger
}

// InitLogger 初始化日志
func InitLogger(logfilePath string) (*Log, error) {
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
		return nil, err
	}

	// 创建日志实例

	writer := zapcore.AddSync(logFile)
	consoleWriter := zapcore.AddSync(os.Stdout)

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "idb",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stack",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		zapcore.NewMultiWriteSyncer(consoleWriter, writer),
		zap.InfoLevel,
	)

	logger := zap.New(core)

	return &Log{Logger: logger}, nil
}

// Info 记录信息日志
func (l *Log) Info(format string, args ...interface{}) {
	l.Logger.Info(fmt.Sprintf(format, args...))
	l.Logger.Sync()
}

// Error 记录错误日志
func (l *Log) Error(format string, args ...interface{}) {
	l.Logger.Error(fmt.Sprintf(format, args...))
	l.Logger.Sync()
}

package log

import (
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sensdata/idb/core/utils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Log 支持热切换文件输出的 logger
type Log struct {
	logger     atomic.Value // *zap.Logger, 用于无锁读取
	logFile    *os.File     // 当前打开的文件句柄，用于手动关闭
	mu         sync.Mutex   // 仅在 Flush/Close 时使用
	installDir string       // 安装目录，用于热切换日志文件
	fileName   string       // 日志文件名，用于热切换日志文件
}

// InitLogger 初始化日志文件和 zap 实例
func InitLogger(installDir, fileName string) (*Log, error) {
	logfilePath := filepath.Join(installDir, fileName)

	if err := utils.EnsureFile(logfilePath); err != nil {
		return nil, err
	}

	logFile, err := os.OpenFile(logfilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	l := &Log{
		logFile:    logFile,
		installDir: installDir,
		fileName:   fileName,
	}

	logger := l.newZapLogger(logFile)
	l.logger.Store(logger)

	// 启动后台自动 sync 协程，保障日志落盘
	go func() {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			curLogger := l.logger.Load().(*zap.Logger)
			if curLogger != nil {
				_ = curLogger.Sync()
			}
		}
	}()

	return l, nil
}

// newZapLogger 创建 zap 实例
func (l *Log) newZapLogger(logFile *os.File) *zap.Logger {
	writer := zapcore.AddSync(logFile)
	console := zapcore.AddSync(os.Stdout)

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
		zapcore.NewMultiWriteSyncer(console, writer),
		zap.InfoLevel,
	)

	return zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
}

// GetLogger 获取当前 zap.Logger 实例（用于自定义日志记录）
func (l *Log) GetLogger() *zap.Logger {
	return l.logger.Load().(*zap.Logger)
}

// Info 打印 Info 级别日志
func (l *Log) Info(format string, args ...interface{}) {
	logger := l.logger.Load().(*zap.Logger)
	if logger != nil {
		logger.Sugar().Infof(format, args...)
	}
}

// Warn 打印 Warn 级别日志
func (l *Log) Warn(format string, args ...interface{}) {
	logger := l.logger.Load().(*zap.Logger)
	if logger != nil {
		logger.Sugar().Warnf(format, args...)
	}
}

// Error 打印 Error 级别日志
func (l *Log) Error(format string, args ...interface{}) {
	logger := l.logger.Load().(*zap.Logger)
	if logger != nil {
		logger.Sugar().Errorf(format, args...)
	}
}

// Flush 重新打开日志文件并切换句柄（用于 logrotate 后执行）
func (l *Log) Flush() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	oldLogger := l.logger.Load().(*zap.Logger)
	oldFile := l.logFile

	newFilePath := filepath.Join(l.installDir, l.fileName)
	newFile, err := os.OpenFile(newFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	newLogger := l.newZapLogger(newFile)
	l.logger.Store(newLogger)
	l.logFile = newFile

	if oldLogger != nil {
		_ = oldLogger.Sync() // 刷新旧日志缓冲
	}
	if oldFile != nil {
		_ = oldFile.Close() // 关闭旧文件句柄，释放 FD
	}

	return nil
}

// Close 手动关闭日志（程序退出时）
func (l *Log) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	logger := l.logger.Load().(*zap.Logger)
	if logger != nil {
		_ = logger.Sync()
	}

	if l.logFile != nil {
		err := l.logFile.Close()
		l.logFile = nil
		return err
	}
	return nil
}

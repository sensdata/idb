package log

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sensdata/idb/core/utils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	defaultMaxSizeMB  = 50
	defaultMaxBackups = 5
	defaultMaxAgeDays = 14
)

// Log 支持热切换文件输出的 logger
type Log struct {
	logger            atomic.Value // *zap.Logger, 用于无锁读取
	logFile           *os.File     // 主日志文件句柄，用于手动关闭
	warnErrorLogFile  *os.File     // warn/error 日志文件句柄，用于手动关闭
	mu                sync.Mutex   // 仅在 Flush/Close 时使用
	installDir        string       // 安装目录，用于热切换日志文件
	fileName          string       // 主日志文件名，用于热切换日志文件
	warnErrorFileName string       // warn/error 日志文件名，用于热切换日志文件
}

// InitLogger 初始化日志文件和 zap 实例
func InitLogger(installDir, fileName string) (*Log, error) {
	logfilePath := filepath.Join(installDir, fileName)
	warnErrorLogfilePath := filepath.Join(installDir, warnErrorFileName(fileName))

	if err := utils.EnsureFile(logfilePath); err != nil {
		return nil, err
	}
	if err := utils.EnsureFile(warnErrorLogfilePath); err != nil {
		return nil, err
	}

	logFile, err := os.OpenFile(logfilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	warnErrorLogFile, err := os.OpenFile(warnErrorLogfilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		_ = logFile.Close()
		return nil, err
	}

	l := &Log{
		logFile:           logFile,
		warnErrorLogFile:  warnErrorLogFile,
		installDir:        installDir,
		fileName:          fileName,
		warnErrorFileName: warnErrorFileName(fileName),
	}

	logger := l.newZapLogger(logFile, warnErrorLogFile)
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
func (l *Log) newZapLogger(logFile *os.File, warnErrorLogFile *os.File) *zap.Logger {
	rotatingWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   logFile.Name(),
		MaxSize:    defaultMaxSizeMB,
		MaxBackups: defaultMaxBackups,
		MaxAge:     defaultMaxAgeDays,
		Compress:   true,
	})
	warnErrorRotatingWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   warnErrorLogFile.Name(),
		MaxSize:    defaultMaxSizeMB,
		MaxBackups: defaultMaxBackups,
		MaxAge:     defaultMaxAgeDays,
		Compress:   true,
	})
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

	consoleCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		console,
		zap.InfoLevel,
	)

	fileCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		rotatingWriter,
		zap.InfoLevel,
	)

	warnErrorFileCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		warnErrorRotatingWriter,
		zap.WarnLevel,
	)

	return zap.New(zapcore.NewTee(consoleCore, fileCore, warnErrorFileCore), zap.AddCaller(), zap.AddCallerSkip(1))
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
	oldWarnErrorFile := l.warnErrorLogFile

	newFilePath := filepath.Join(l.installDir, l.fileName)
	newFile, err := os.OpenFile(newFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	newWarnErrorFilePath := filepath.Join(l.installDir, l.warnErrorFileName)
	newWarnErrorFile, err := os.OpenFile(newWarnErrorFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		_ = newFile.Close()
		return err
	}

	newLogger := l.newZapLogger(newFile, newWarnErrorFile)
	l.logger.Store(newLogger)
	l.logFile = newFile
	l.warnErrorLogFile = newWarnErrorFile

	if oldLogger != nil {
		_ = oldLogger.Sync() // 刷新旧日志缓冲
	}
	if oldFile != nil {
		_ = oldFile.Close() // 关闭旧文件句柄，释放 FD
	}
	if oldWarnErrorFile != nil {
		_ = oldWarnErrorFile.Close()
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
		if l.warnErrorLogFile != nil {
			_ = l.warnErrorLogFile.Close()
			l.warnErrorLogFile = nil
		}
		return err
	}
	if l.warnErrorLogFile != nil {
		err := l.warnErrorLogFile.Close()
		l.warnErrorLogFile = nil
		return err
	}
	return nil
}

func warnErrorFileName(fileName string) string {
	ext := filepath.Ext(fileName)
	if ext == "" {
		return fileName + ".warn-error"
	}

	base := strings.TrimSuffix(fileName, ext)
	return base + ".warn-error" + ext
}

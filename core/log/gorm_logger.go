package log

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

// GormLogger 实现gorm.io/gorm/logger.Writer接口
type GormLogger struct {
	log *Log // 直接持有Log实例，利用其内部的atomic.Value机制自动获取最新的logger
}

// NewGormLogger 创建新的GormLogger，接受Log实例
func NewGormLogger(log *Log) *GormLogger {
	return &GormLogger{log: log}
}

// getLogger 获取最新的zap.Logger实例
func (l *GormLogger) getLogger() *zap.Logger {
	if l.log != nil {
		return l.log.GetLogger()
	}
	// 默认返回nop logger避免nil panic
	return zap.NewNop()
}

// LogMode 设置日志级别（实现logger.Interface接口）
func (l *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	// 为不同日志级别创建新的GormLogger
	newLogger := *l
	switch level {
	case logger.Silent:
		// 静默模式下创建一个特殊的Log实例，总是返回nop logger
		silentLog := &Log{}
		// 预存储一个nop logger
		nopLogger := zap.NewNop()
		silentLog.logger.Store(nopLogger)
		newLogger.log = silentLog
	default:
		// 其他模式保持原样，Log实例内部会处理日志级别的调整
	}
	return &newLogger
}

// Info 记录信息日志（实现logger.Interface接口）
func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	l.getLogger().Sugar().Infof(msg, data...)
}

// Warn 记录警告日志（实现logger.Interface接口）
func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	l.getLogger().Sugar().Warnf(msg, data...)
}

// Error 记录错误日志（实现logger.Interface接口）
func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	l.getLogger().Sugar().Errorf(msg, data...)
}

// Trace 记录SQL跟踪日志（实现logger.Interface接口）
func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	logger := l.getLogger()
	elapsed := time.Since(begin)
	sql, rows := fc()
	if err != nil {
		logger.Sugar().Errorf("SQL: %s | Elapsed: %s | Rows: %d | Error: %s", sql, elapsed, rows, err)
	} else {
		logger.Sugar().Infof("SQL: %s | Elapsed: %s | Rows: %d", sql, elapsed, rows)
	}
}

// Printf 格式化打印日志（实现logger.Writer接口）
func (l *GormLogger) Printf(format string, args ...interface{}) {
	logger := l.getLogger()
	// 根据日志内容确定日志级别
	var logLevel zapcore.Level
	if strings.Contains(format, "error") {
		logLevel = zap.ErrorLevel
	} else if strings.Contains(format, "panic") {
		logLevel = zap.PanicLevel
	} else if strings.Contains(format, "warn") {
		logLevel = zap.WarnLevel
	} else {
		logLevel = zap.InfoLevel
	}

	// 格式化消息
	msg := fmt.Sprintf(format, args...)

	// 使用适当的Zap日志函数记录
	switch logLevel {
	case zap.ErrorLevel:
		logger.Error(msg)
	case zap.WarnLevel:
		logger.Warn(msg)
	case zap.InfoLevel:
		logger.Info(msg)
	case zap.DebugLevel:
		logger.Debug(msg)
	case zap.PanicLevel:
		logger.Panic(msg)
	default:
		logger.Info(msg)
	}
}

// Print 打印日志（实现logger.Writer接口）
func (l *GormLogger) Print(v ...interface{}) {
	logger := l.getLogger()
	if len(v) > 1 {
		level, ok1 := v[0].(string)
		if ok1 && level == "sql" {
			source := "unknown"
			if len(v) > 1 {
				source = utils.ToString(v[1])
			}
			logger.Info("SQL", zap.String("Source", source))
		} else {
			// 通用打印
			logger.Info(fmt.Sprint(v...))
		}
	} else if len(v) == 1 {
		logger.Info(fmt.Sprint(v...))
	}
}

// Println 打印带换行的日志（实现logger.Writer接口）
func (l *GormLogger) Println(v ...interface{}) {
	l.Print(v...)
}

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

type GormLogger struct {
	Logger *zap.Logger
}

func NewGormLogger(zapLogger *zap.Logger) *GormLogger {
	return &GormLogger{Logger: zapLogger}
}

func (l *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	// Set the log level based on GORM's log level
	newLogger := *l
	switch level {
	case logger.Silent:
		newLogger.Logger = zap.NewNop()
	case logger.Error:
		newLogger.Logger = l.Logger.WithOptions(zap.IncreaseLevel(zap.ErrorLevel))
	case logger.Warn:
		newLogger.Logger = l.Logger.WithOptions(zap.IncreaseLevel(zap.WarnLevel))
	case logger.Info:
		newLogger.Logger = l.Logger.WithOptions(zap.IncreaseLevel(zap.InfoLevel))
	}
	return &newLogger
}

func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	l.Logger.Sugar().Infof(msg, data...)
}

func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	l.Logger.Sugar().Warnf(msg, data...)
}

func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	l.Logger.Sugar().Errorf(msg, data...)
}

func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()
	if err != nil {
		l.Logger.Sugar().Errorf("SQL: %s | Elapsed: %s | Rows: %d | Error: %s", sql, elapsed, rows, err)
	} else {
		l.Logger.Sugar().Infof("SQL: %s | Elapsed: %s | Rows: %d", sql, elapsed, rows)
	}
}

func (l *GormLogger) Printf(format string, args ...interface{}) {
	// Use a custom log level based on the log level defined in GORM
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

	// Format the message using fmt.Sprintf
	msg := fmt.Sprintf(format, args...)

	// Log using the appropriate Zap logging function
	switch logLevel {
	case zap.ErrorLevel:
		l.Logger.Error(msg)
	case zap.WarnLevel:
		l.Logger.Warn(msg)
	case zap.InfoLevel:
		l.Logger.Info(msg)
	case zap.DebugLevel:
		l.Logger.Debug(msg)
	case zap.PanicLevel:
		l.Logger.Panic(msg)
	default:
		l.Logger.Info(msg)
	}
}

func (l *GormLogger) Print(v ...interface{}) {
	if len(v) > 1 {
		level := v[0]
		source := v[1]
		if level == "sql" {
			l.Logger.Info("SQL",
				zap.String("Source", utils.ToString(source)),
			)
		}
	}
}

func (l *GormLogger) Println(v ...interface{}) {
	l.Print(v...)
}

package types

import "errors"

var (
	// 任务相关错误
	ErrTaskNotFound  = errors.New("task not found")
	ErrInvalidStatus = errors.New("invalid task status")
	ErrTaskExists    = errors.New("task already exists")
	ErrTaskExpired   = errors.New("task expired")

	// 读写器相关错误
	ErrReaderClosed = errors.New("reader is closed")
	ErrWriterClosed = errors.New("writer is closed")
	ErrReaderBusy   = errors.New("reader is busy")
	ErrWriterBusy   = errors.New("writer is busy")

	// 日志相关错误
	ErrInvalidLogLevel = errors.New("invalid log level")
	ErrLogSizeExceeded = errors.New("log size exceeded")
	ErrLogNotFound     = errors.New("log file not found")

	// 系统相关错误
	ErrSystemBusy    = errors.New("system is busy")
	ErrConfigInvalid = errors.New("invalid configuration")
	ErrCleanupFailed = errors.New("cleanup failed")
)

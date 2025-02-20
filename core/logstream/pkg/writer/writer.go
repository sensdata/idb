package writer

import (
	"github.com/sensdata/idb/core/logstream/pkg/types"
)

// Writer 日志写入器接口
type Writer interface {
	// Write 写入格式化日志
	Write(level types.LogLevel, message string, metadata map[string]string) error

	// WriteRaw 写入原始数据
	WriteRaw(data []byte) error

	// Close 关闭写入器
	Close() error
}

package reader

import (
	"github.com/sensdata/idb/core/logstream/pkg/types"
)

// Reader 日志读取器接口
type Reader interface {
	// Read 从指定位置读取日志
	Read(offset int64) ([]byte, error)

	// Follow 持续读取新日志
	Follow(offset int64, whence int) (<-chan []byte, error)

	// Follow 日志结构
	FollowEntry() (<-chan types.LogEntry, error)

	// Close 关闭读取器
	Close() error

	// Open 
	Open() error
}

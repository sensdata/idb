package adapters

import (
	"bytes"
	"fmt"
	"io"
	"sync"

	"github.com/sensdata/idb/core/logstream/pkg/types"
)

type BufferReader struct {
	mu     sync.RWMutex
	buffer *bytes.Buffer
	subs   []chan []byte
	closed bool
}

// NewBufferReader 使用已存在的 buffer 创建 reader
func NewBufferReader(buffer *bytes.Buffer) *BufferReader {
	if buffer == nil {
		buffer = bytes.NewBuffer(nil)
	}
	return &BufferReader{
		buffer: buffer,
		subs:   make([]chan []byte, 0),
	}
}

func (r *BufferReader) Read(offset int64) ([]byte, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.closed {
		return nil, fmt.Errorf("reader is closed")
	}

	if offset >= int64(r.buffer.Len()) {
		return nil, io.EOF
	}

	data := r.buffer.Bytes()[offset:]
	return data, nil
}

func (r *BufferReader) Follow(offset int64, whence int) (<-chan []byte, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil, fmt.Errorf("reader is closed")
	}

	ch := make(chan []byte, 100)
	r.subs = append(r.subs, ch)

	// 发送现有内容
	if r.buffer.Len() > 0 {
		ch <- r.buffer.Bytes()
	}

	return ch, nil
}

func (r *BufferReader) FollowEntry() (<-chan types.LogEntry, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	ch := make(chan types.LogEntry)
	return ch, nil
}

func (r *BufferReader) Write(data []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return fmt.Errorf("reader is closed")
	}

	r.buffer.Write(data)

	// 通知所有订阅者
	for _, ch := range r.subs {
		select {
		case ch <- data:
		default:
			// 如果通道已满，丢弃日志
		}
	}

	return nil
}

func (r *BufferReader) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil
	}

	r.closed = true
	for _, ch := range r.subs {
		close(ch)
	}
	r.subs = nil
	r.buffer.Reset()

	return nil
}

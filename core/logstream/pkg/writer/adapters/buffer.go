package adapters

import (
	"bytes"
	"encoding/json"
	"sync"
	"time"

	"github.com/sensdata/idb/core/logstream/pkg/types"
)

type BufferWriter struct {
	mu     sync.Mutex
	buffer *bytes.Buffer
}

func NewBufferWriter() *BufferWriter {
	return &BufferWriter{
		buffer: bytes.NewBuffer(nil),
	}
}

func (w *BufferWriter) Write(level types.LogLevel, message string, metadata map[string]string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	entry := types.LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Metadata:  metadata,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	w.buffer.Write(append(data, '\n'))
	return nil
}

func (w *BufferWriter) WriteRaw(data []byte) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if len(data) > 0 {
		if data[len(data)-1] != '\n' {
			data = append(data, '\n')
		}
		w.buffer.Write(data)
	}
	return nil
}

func (w *BufferWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.buffer.Reset()
	return nil
}

func (w *BufferWriter) Bytes() []byte {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.buffer.Bytes()
}

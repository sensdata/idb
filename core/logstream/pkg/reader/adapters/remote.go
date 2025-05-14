package adapters

import (
	"fmt"
	"sync"

	"github.com/sensdata/idb/core/logstream/internal/config"
)

type RemoteReader struct {
	mu     sync.Mutex
	logCh  chan []byte
	done   chan struct{}
	closed bool
	config *config.Config
}

func NewRemoteReader(filePath string, cfg *config.Config) *RemoteReader {
	return &RemoteReader{
		logCh:  make(chan []byte, cfg.MaxFollowBuffer),
		done:   make(chan struct{}),
		config: cfg,
	}
}

func (r *RemoteReader) Read(offset int64) ([]byte, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil, fmt.Errorf("reader is closed")
	}
	return []byte{}, nil
}

func (r *RemoteReader) Follow(offset int64, whence int) (<-chan []byte, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil, fmt.Errorf("reader is closed")
	}

	return r.logCh, nil
}

func (r *RemoteReader) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil
	}

	r.closed = true
	close(r.done)

	return nil
}

func (r *RemoteReader) Open() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.closed {
		return nil
	}

	r.logCh = make(chan []byte, r.config.MaxFollowBuffer)
	r.done = make(chan struct{})
	r.closed = false

	return nil
}

func (r *RemoteReader) SendLog(data []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return fmt.Errorf("reader is closed")
	}

	select {
	case r.logCh <- data:
		return nil
	case <-r.done:
		return fmt.Errorf("reader is done")
	default:
		// 通道已满时，直接覆盖最旧的消息
		<-r.logCh
		r.logCh <- data
		return nil
	}
}

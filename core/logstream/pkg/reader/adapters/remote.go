package adapters

import (
	"fmt"
	"sync"

	"github.com/sensdata/idb/core/logstream/internal/config"
	"github.com/sensdata/idb/core/logstream/pkg/types"
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

func (r *RemoteReader) FollowEntry() (<-chan types.LogEntry, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil, fmt.Errorf("reader is closed")
	}
	ch := make(chan types.LogEntry, r.config.MaxFollowBuffer)
	return ch, nil
}

func (r *RemoteReader) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil
	}

	r.closed = true
	close(r.done)

	// 发送停止追踪请求
	// stopMsg, err := message.CreateLogStreamMessage(
	// 	utils.GenerateMsgId(),
	// 	message.LogStreamStop,
	// 	r.taskID,
	// 	r.filePath,
	// 	"",
	// 	"",
	// )
	// if err != nil {
	// 	return fmt.Errorf("create stop message failed: %w", err)
	// }

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

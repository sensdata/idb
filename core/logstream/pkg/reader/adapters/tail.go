package adapters

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/nxadm/tail"
	"github.com/sensdata/idb/core/logstream/internal/config"
	"github.com/sensdata/idb/core/logstream/pkg/types"
)

type TailReader struct {
	mu       sync.Mutex
	file     *os.File
	tail     *tail.Tail
	filePath string
	closed   bool
	config   *config.Config
}

func NewTailReader(filePath string, cfg *config.Config) (*TailReader, error) {
	file, err := os.OpenFile(filePath, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, fmt.Errorf("open file failed: %v", err)
	}
	// 创建配置的副本
	var conf config.Config
	if cfg != nil {
		conf = *cfg
	} else {
		conf = *config.DefaultConfig()
	}

	return &TailReader{
		file:     file,
		filePath: filePath,
		config:   &conf,
	}, nil
}

func (r *TailReader) Read(offset int64) ([]byte, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil, fmt.Errorf("reader is closed")
	}

	if _, err := r.file.Seek(offset, io.SeekStart); err != nil {
		return nil, fmt.Errorf("seek file failed: %v", err)
	}

	reader := bufio.NewReaderSize(r.file, r.config.ReadBufferSize)
	line, err := reader.ReadBytes('\n')
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("read file failed: %v", err)
	}

	return line, nil
}

func (r *TailReader) Follow(offset int64, whence int) (<-chan []byte, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil, fmt.Errorf("reader is closed")
	}

	// 如果已经有tail实例，先停止它
	if r.tail != nil {
		r.tail.Stop()
		r.tail.Cleanup()
	}

	config := tail.Config{
		Follow:    true,
		ReOpen:    true,
		MustExist: false,
		Poll:      true,
		Location: &tail.SeekInfo{
			Offset: offset,
			Whence: whence,
		},
	}

	t, err := tail.TailFile(r.filePath, config)
	if err != nil {
		return nil, fmt.Errorf("tail file failed: %v", err)
	}

	r.tail = t
	ch := make(chan []byte, r.config.MaxFollowBuffer)

	go func() {
		defer close(ch)
		for line := range t.Lines {
			if line.Err != nil {
				continue
			}
			select {
			case ch <- []byte(line.Text + "\n"):
			default:
				// 如果通道已满，丢弃日志
			}
		}
	}()

	return ch, nil
}

func (r *TailReader) FollowEntry() (<-chan types.LogEntry, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil, fmt.Errorf("reader is closed")
	}

	if r.tail != nil {
		return nil, fmt.Errorf("already following")
	}

	config := tail.Config{
		Follow:    true,
		ReOpen:    true,
		MustExist: false,
		Poll:      true,
		Location: &tail.SeekInfo{
			Offset: 0,
			Whence: io.SeekStart,
		},
	}

	t, err := tail.TailFile(r.filePath, config)
	if err != nil {
		return nil, fmt.Errorf("tail file failed: %v", err)
	}

	r.tail = t
	ch := make(chan types.LogEntry, r.config.MaxFollowBuffer)

	go func() {
		defer close(ch)
		for line := range t.Lines {
			if line.Err != nil {
				continue
			}

			var entry types.LogEntry
			if err := json.Unmarshal([]byte(line.Text), &entry); err != nil {
				// 如果解析失败，跳过这一行
				continue
			}

			select {
			case ch <- entry:
			default:
				// 如果通道已满，丢弃日志
			}
		}
	}()

	return ch, nil
}

func (r *TailReader) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil
	}

	r.closed = true

	var errs []error

	if r.tail != nil {
		r.tail.Stop()
		r.tail.Cleanup()
		r.tail = nil
	}

	if r.file != nil {
		if err := r.file.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close file failed: %v", err))
		}
		r.file = nil
	}

	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

func (r *TailReader) Open() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.closed {
		return nil
	}

	file, err := os.OpenFile(r.filePath, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("open file failed: %v", err)
	}

	r.file = file
	r.closed = false
	return nil
}

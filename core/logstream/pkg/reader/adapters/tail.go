package adapters

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/nxadm/tail"
	"github.com/sensdata/idb/core/logstream/internal/config"
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

	return &TailReader{
		file:     file,
		filePath: filePath,
		config:   cfg,
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

func (r *TailReader) Follow() (<-chan []byte, error) {
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

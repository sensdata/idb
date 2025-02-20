package adapters

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/sensdata/idb/core/logstream/pkg/types"
)

type FileWriter struct {
	mu       sync.Mutex
	file     *os.File
	writer   *bufio.Writer
	filePath string
}

func NewFileWriter(filePath string, bufferSize int) (*FileWriter, error) {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create directory failed: %v", err)
	}

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("open file failed: %v", err)
	}

	return &FileWriter{
		file:     file,
		writer:   bufio.NewWriterSize(file, bufferSize),
		filePath: filePath,
	}, nil
}

func (w *FileWriter) Write(level types.LogLevel, message string, metadata map[string]string) error {
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
		return fmt.Errorf("marshal log entry failed: %v", err)
	}

	if _, err := w.writer.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("write log failed: %v", err)
	}

	return w.writer.Flush()
}

func (w *FileWriter) WriteRaw(data []byte) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if len(data) == 0 {
		return nil
	}

	if data[len(data)-1] != '\n' {
		data = append(data, '\n')
	}

	if _, err := w.writer.Write(data); err != nil {
		return fmt.Errorf("write raw data failed: %v", err)
	}

	return w.writer.Flush()
}

func (w *FileWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if err := w.writer.Flush(); err != nil {
		return fmt.Errorf("flush writer failed: %v", err)
	}

	if err := w.file.Close(); err != nil {
		return fmt.Errorf("close file failed: %v", err)
	}

	return nil
}

package adapters

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"time"
)

type ServiceLogReader struct {
	mu      sync.RWMutex
	service string
	since   string
	number  string
	follow  bool
	closed  bool
	cmd     *exec.Cmd
	stdout  io.ReadCloser
}

// 构造函数
func NewServiceLogReader(service, since, number string, follow bool) (*ServiceLogReader, error) {
	return &ServiceLogReader{
		service: service,
		since:   since,
		number:  number,
		follow:  follow,
	}, nil
}

// 构建命令
func (r *ServiceLogReader) buildCmd() *exec.Cmd {
	var cmdName string
	var args []string
	cmdName = "journalctl"
	args = []string{"-u", r.service}
	if r.since != "" && r.since != "all" {
		args = append(args, "--since", r.since)
	}
	if r.number != "" && r.number != "all" {
		args = append(args, "-n", r.number)
	}
	if r.follow {
		args = append(args, "-f")
	}
	// 加上 --no-pager 避免分页
	args = append(args, "--no-pager")
	return exec.Command(cmdName, args...)
}

// 一次性读取日志
func (r *ServiceLogReader) Read(offset int64) ([]byte, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return nil, fmt.Errorf("reader is closed")
	}
	cmd := r.buildCmd()
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("exec log command failed: %v", err)
	}
	return out.Bytes(), nil
}

// 持续读取日志
func (r *ServiceLogReader) Follow(offset int64, whence int) (<-chan []byte, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return nil, fmt.Errorf("reader is closed")
	}
	cmd := r.buildCmd()
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("get stdout pipe failed: %v", err)
	}
	cmd.Stderr = cmd.Stdout
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start log command failed: %v", err)
	}
	r.cmd = cmd
	r.stdout = stdout

	ch := make(chan []byte, 100)
	go func() {
		defer close(ch)
		defer func() {
			r.mu.Lock()
			r.closed = true
			r.mu.Unlock()
			if r.cmd != nil && r.cmd.Process != nil {
				_ = r.cmd.Process.Kill()
			}
		}()
		reader := bufio.NewReader(stdout)
		for {
			line, err := reader.ReadBytes('\n')
			if len(line) > 0 {
				select {
				case ch <- line:
				default:
					// 丢弃
				}
			}
			if err != nil {
				if err == io.EOF {
					return
				}
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()
	return ch, nil
}

// 关闭读取器
func (r *ServiceLogReader) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return nil
	}
	r.closed = true
	if r.cmd != nil && r.cmd.Process != nil {
		return r.cmd.Process.Kill()
	}
	if r.stdout != nil {
		_ = r.stdout.Close()
	}
	return nil
}

// 重新建立命令管道
func (r *ServiceLogReader) Open() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.closed {
		return nil
	}
	cmd := r.buildCmd()
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("get stdout pipe failed: %v", err)
	}
	cmd.Stderr = cmd.Stdout
	r.cmd = cmd
	r.stdout = stdout
	r.closed = false
	return nil
}

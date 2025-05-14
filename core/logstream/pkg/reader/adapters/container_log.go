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

type ContainerLogReader struct {
	mu            sync.Mutex
	containerType string // "docker" 或 "compose"
	container     string // 容器ID/名称 或 compose文件路径
	since         string // all, 24h, 4h, 1h, 10m
	tail          string
	follow        bool
	closed        bool
	cmd           *exec.Cmd
	stdout        io.ReadCloser
}

// 构造函数
func NewContainerLogReader(containerType, container, since, tail string, follow bool) (*ContainerLogReader, error) {
	return &ContainerLogReader{
		containerType: containerType,
		container:     container,
		since:         since,
		tail:          tail,
		follow:        follow,
	}, nil
}

// 构建命令
func (r *ContainerLogReader) buildCmd(follow bool) *exec.Cmd {
	var cmdName string
	var args []string
	if r.containerType == "compose" {
		cmdName = "docker-compose"
		args = []string{"-f", r.container, "logs"}
	} else {
		cmdName = "docker"
		args = []string{"logs", r.container}
	}
	if r.tail != "" && r.tail != "0" {
		args = append(args, "--tail", r.tail)
	}
	if r.since != "" && r.since != "all" {
		args = append(args, "--since", r.since)
	}
	if follow {
		args = append(args, "-f")
	}
	return exec.Command(cmdName, args...)
}

// 一次性读取日志
func (r *ContainerLogReader) Read(offset int64) ([]byte, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return nil, fmt.Errorf("reader is closed")
	}
	cmd := r.buildCmd(false)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("exec log command failed: %v", err)
	}
	return out.Bytes(), nil
}

// 持续读取日志
func (r *ContainerLogReader) Follow(offset int64, whence int) (<-chan []byte, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return nil, fmt.Errorf("reader is closed")
	}
	cmd := r.buildCmd(true)
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
func (r *ContainerLogReader) Close() error {
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
func (r *ContainerLogReader) Open() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.closed {
		return nil
	}
	cmd := r.buildCmd(r.follow)
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

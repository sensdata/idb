package conn

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/global"
)

const (
	heartbeatInterval = 30 * time.Second // 心跳间隔
	connectionTimeout = 2 * time.Hour    // 连接超时时间
	maxBufferSize     = 1024 * 1024      // 日志缓冲区大小限制（1MB）
)

type ITaskService interface {
	HandleTaskLogStream(c *gin.Context) error
}

type TaskService struct{}

func NewTaskService() *TaskService {
	return &TaskService{}
}

func (s *TaskService) HandleTaskLogStream(c *gin.Context) error {
	ls := global.LogStream
	taskID := c.Param("taskId")
	if taskID == "" {
		return errors.New("invalid task ID")
	}

	// 获取任务状态监听器
	watcher, err := ls.GetTaskWatcher(taskID)
	if err != nil {
		global.LOG.Error("get task watcher failed: %v", err)
		return fmt.Errorf("get task watcher failed: %w", err)
	}
	defer watcher.Close()

	reader, err := ls.GetReader(taskID)
	if err != nil {
		global.LOG.Error("get reader failed: %v", err)
		return fmt.Errorf("get reader failed: %w", err)
	}
	defer reader.Close()

	logCh, err := reader.Follow()
	if err != nil {
		global.LOG.Error("follow log failed: %v", err)
		return fmt.Errorf("follow log failed: %w", err)
	}

	// 获取状态监听通道
	statusCh, err := watcher.WatchStatus()
	if err != nil {
		global.LOG.Error("watch status failed: %v", err)
		return fmt.Errorf("watch status failed: %w", err)
	}

	clientGone := c.Request.Context().Done()
	heartbeat := time.NewTicker(heartbeatInterval)
	defer heartbeat.Stop()
	timeout := time.NewTimer(connectionTimeout)
	defer timeout.Stop()

	// 添加一个 done 通道用于控制 goroutine 退出
	done := make(chan struct{})
	defer close(done)

	bufferedCh := make(chan []byte, maxBufferSize/1024)
	defer close(bufferedCh)

	// 启动日志处理协程
	go func() {
		defer func() {
			if r := recover(); r != nil {
				global.LOG.Error("log processor panic: %v", r)
			}
		}()

		for {
			select {
			case msg, ok := <-logCh:
				if !ok {
					return
				}
				select {
				case bufferedCh <- msg:
				default:
					global.LOG.Warn("Log buffer full, dropping messages for task: %s", taskID)
				}
			case <-clientGone:
				return
			case <-done:
				return
			}
		}
	}()

	// 使用 context 来控制超时和客户端断开
	ctx, cancel := context.WithTimeout(c.Request.Context(), connectionTimeout)
	defer cancel()

	c.Stream(func(w io.Writer) bool {
		select {
		case msg, ok := <-bufferedCh:
			if !ok {
				return false
			}
			c.SSEvent("log", string(msg))
			return true
		case status, ok := <-statusCh:
			if !ok {
				return false
			}
			// 发送状态变更事件并检查错误
			c.SSEvent("status", status)
			return true
		case <-heartbeat.C:
			c.SSEvent("heartbeat", time.Now().Unix())
			return true
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				c.SSEvent("error", "Connection timeout")
			}
			return false
		}
	})

	return nil
}

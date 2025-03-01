package conn

import (
	"context"
	"errors"
	"fmt"
	"net/http"
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

	// 获取任务状态监听器
	watcher, err := ls.GetTaskWatcher(taskID)
	if err != nil {
		global.LOG.Error("get task watcher failed: %v", err)
		return fmt.Errorf("get task watcher failed: %w", err)
	}
	defer watcher.Close()

	// 获取状态监听通道
	statusCh, err := watcher.WatchStatus()
	if err != nil {
		global.LOG.Error("watch status failed: %v", err)
		return fmt.Errorf("watch status failed: %w", err)
	}

	// 使用 context 来控制超时和客户端断开
	ctx, cancel := context.WithTimeout(c.Request.Context(), connectionTimeout)
	defer cancel()

	heartbeat := time.NewTicker(heartbeatInterval)
	defer heartbeat.Stop()

	// 添加一个 done 通道用于控制 goroutine 退出
	done := make(chan struct{})
	defer close(done)

	// 设置 SSE 响应头
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	// 创建一个缓冲通道来处理日志
	bufferCh := make(chan []byte, 100)
	defer close(bufferCh)

	// 启动一个 goroutine 来处理日志缓冲
	go func() {
		for {
			select {
			case msg := <-logCh:
				select {
				case bufferCh <- msg:
				default:
					// 如果缓冲区满了，丢弃最旧的消息
					<-bufferCh
					bufferCh <- msg
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		return fmt.Errorf("streaming not supported")
	}

	for {
		select {
		case msg := <-bufferCh:
			c.SSEvent("log", string(msg))
			flusher.Flush()
		case status := <-statusCh:
			c.SSEvent("status", status)
			flusher.Flush()
		case <-heartbeat.C:
			c.SSEvent("heartbeat", time.Now().Unix())
			flusher.Flush()
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				c.SSEvent("error", "Connection timeout")
			} else {
				c.SSEvent("error", "Connection closed")
			}
			flusher.Flush()
			return nil
		}
	}
}

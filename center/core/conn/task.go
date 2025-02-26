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

	c.Stream(func(w io.Writer) bool {
		select {
		case msg := <-logCh:
			global.LOG.Info("task logCh: %s", string(msg))
			c.SSEvent("log", string(msg))
			return true
		case status := <-statusCh:
			global.LOG.Info("task statusCh: %s", string(status))
			c.SSEvent("status", status)
			return true
		case <-heartbeat.C:
			global.LOG.Info("task heartbeat")
			c.SSEvent("heartbeat", time.Now().Unix())
			return true
		case <-ctx.Done():
			global.LOG.Info("task done")
			if ctx.Err() == context.DeadlineExceeded {
				c.SSEvent("error", "Connection timeout")
			} else {
				c.SSEvent("error", "Connection closed")
			}
			return false
		}
	})

	return nil
}

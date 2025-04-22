package conn

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/logstream/pkg/reader/adapters"
	"github.com/sensdata/idb/core/message"
	"github.com/sensdata/idb/core/utils"
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

	offset, err := strconv.ParseInt(c.Query("offset"), 10, 32)
	if err != nil {
		offset = 0
	}

	var whence = io.SeekStart
	w := c.Query("whence")
	switch w {
	case "start":
		whence = io.SeekStart
		if offset < 0 {
			offset = 0
		}
	case "end":
		whence = io.SeekEnd
		if offset >= 0 {
			offset = -1024
		}
	default:
		whence = io.SeekStart
		offset = 0
	}

	// 获取任务信息
	task, err := ls.GetTask(taskID)
	if err != nil {
		global.LOG.Error("get task failed: %v", err)
		return fmt.Errorf("get task failed: %w", err)
	}

	// 把task的metadata都打印出来
	for k, v := range task.Metadata {
		global.LOG.Info("task metadata: %s=%v", k, v)
	}

	reader, err := ls.GetReader(task.ID)
	if err != nil {
		global.LOG.Error("get reader failed: %v", err)
		return fmt.Errorf("get reader failed: %w", err)
	}
	defer reader.Close()

	// 判断reader是否是 RemoteReader
	_, ok := reader.(*adapters.RemoteReader)
	if ok {
		// 找host
		id, ok := task.Metadata["host"]
		if !ok {
			global.LOG.Error("cannot find host in task metadata")
			return errors.New("invalid host")
		}
		// id 转成 uint
		hostId, ok := id.(uint)
		if !ok {
			global.LOG.Error("invalid host id")
			return errors.New("invalid host id")
		}

		// 找host
		host, err := HostRepo.Get(HostRepo.WithByID(uint(hostId)))
		if err != nil {
			global.LOG.Error("get host failed: %v", err)
			return fmt.Errorf("get host failed: %w", err)
		}
		// 获取agent连接
		conn, err := CENTER.GetAgentConn(&host)
		if err != nil {
			global.LOG.Error("get agent conn failed: %v", err)
			return fmt.Errorf("get agent conn failed: %w", err)
		}

		// 发送开始追踪请求
		startMsg, err := message.CreateLogStreamMessage(
			utils.GenerateMsgId(),
			message.LogStreamStart,
			task.ID,
			task.LogPath,
			offset,
			whence,
			"",
			"",
		)
		if err != nil {
			return fmt.Errorf("create start message failed: %w", err)
		}

		if err := message.SendLogStreamMessage(*conn, startMsg); err != nil {
			return fmt.Errorf("send start message failed: %w", err)
		}
	}

	logCh, err := reader.Follow(offset, whence)
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
			global.LOG.Info("SSE STATUS: %s", status)
			c.SSEvent("status", status)
			flusher.Flush()
		case <-heartbeat.C:
			global.LOG.Info("SSE HEARTBEAT")
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

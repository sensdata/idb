package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/core/conn"
	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/logstream/pkg/reader/adapters"
	"github.com/sensdata/idb/core/logstream/pkg/types"
	"github.com/sensdata/idb/core/message"
	"github.com/sensdata/idb/core/utils"
)

const (
	heartbeatInterval = 30 * time.Second // 心跳间隔
	connectionTimeout = 2 * time.Hour    // 连接超时时间
	maxBufferSize     = 1024 * 1024      // 日志缓冲区大小限制（1MB）
)

type ILogManService interface {
	HandleLogStream(c *gin.Context) error
}

type LogManService struct{}

func NewILogManService() ILogManService {
	return &LogManService{}
}

func (s *LogManService) HandleLogStream(c *gin.Context) error {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		return errors.New("invalid host")
	}

	path := c.Query("path")
	if path == "" {
		return errors.New("invalid path")
	}

	var offset int64
	var whence int
	w := c.Query("whence")
	switch w {
	case "end":
		whence = io.SeekEnd
		offset = 0
	default:
		whence = io.SeekStart
		offset = 0
	}

	// 找host
	host, err := HostRepo.Get(HostRepo.WithByID(uint(hostID)))
	if err != nil {
		global.LOG.Error("get host failed: %v", err)
		return fmt.Errorf("get host failed: %w", err)
	}

	ls := global.LogStream

	// 查找任务
	var task *types.Task
	task, err = ls.GetTaskByLog(path)
	if err != nil {
		global.LOG.Error("get task failed: %v", err)
	}
	if task == nil {
		global.LOG.Info("task not found, creating new task")
		// 创建任务
		metadata := map[string]interface{}{
			"log_path": path,
		}
		// 本机
		if host.IsDefault {
			task, err = ls.CreateTask(types.TaskTypeFile, metadata)
			if err != nil {
				return errors.New("failed to create tail task")
			}
		} else {
			task, err = ls.CreateTask(types.TaskTypeRemote, metadata)
			if err != nil {
				return errors.New("failed to create tail task")
			}
		}
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
		global.LOG.Info("reader is RemoteReader")
		// 获取agent连接
		conn, err := conn.CENTER.GetAgentConn(&host)
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
	global.LOG.Info("follow log success")

	// 获取任务状态监听器
	watcher, err := ls.GetTaskWatcher(task.ID)
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

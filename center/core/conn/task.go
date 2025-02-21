package conn

import (
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

var taskService = NewTaskService()

func (s *TaskService) HandleTaskLogStream(c *gin.Context) error {
	ls := global.LogStream
	taskID := c.Param("taskId")
	if taskID == "" {
		return errors.New("Invalid task ID")
	}

	reader, err := ls.GetReader(taskID)
	if err != nil {
		return fmt.Errorf("get reader failed: %w", err)
	}

	// 确保资源清理
	defer reader.Close()

	logCh, err := reader.Follow()
	if err != nil {
		return fmt.Errorf("follow log failed: %w", err)
	}

	// 检测客户端断开连接
	clientGone := c.Request.Context().Done()

	// 创建心跳定时器
	heartbeat := time.NewTicker(heartbeatInterval)
	defer heartbeat.Stop()

	// 设置连接超时
	timeout := time.NewTimer(connectionTimeout)
	defer timeout.Stop()

	// 创建一个带缓冲的通道用于日志传输
	bufferedCh := make(chan []byte, maxBufferSize/1024) // 假设每条日志平均1KB
	defer close(bufferedCh)

	// 启动日志处理协程
	go func() {
		for {
			select {
			case msg, ok := <-logCh:
				if !ok {
					return
				}
				// 如果缓冲区满，丢弃旧的消息
				select {
				case bufferedCh <- msg:
				default:
					global.LOG.Warn("Log buffer full, dropping messages for task: %s", taskID)
				}
			case <-clientGone:
				return
			}
		}
	}()

	c.Stream(func(w io.Writer) bool {
		select {
		case msg, ok := <-bufferedCh:
			if !ok {
				return false
			}
			c.SSEvent("log", string(msg))
			return true
		case <-heartbeat.C:
			// 发送心跳消息
			c.SSEvent("heartbeat", time.Now().Unix())
			return true
		case <-timeout.C:
			// 连接超时
			c.SSEvent("error", "Connection timeout")
			return false
		case <-clientGone:
			return false
		}
	})

	return nil
}

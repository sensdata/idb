package service

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
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
	heartbeatInterval = 10 * time.Second // 心跳间隔
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

	// 创建任务
	var task *types.Task
	metadata := map[string]interface{}{
		"log_path": path,
	}
	// 本机（一些非依赖agent的系统交互过程日志，）
	if strings.Contains(path, "/idb/data/logstream/logs") {
		task, err = global.LogStream.CreateTask(types.TaskTypeFile, metadata)
		if err != nil {
			return errors.New("failed to create tail task")
		}
	} else {
		task, err = global.LogStream.CreateTask(types.TaskTypeRemote, metadata)
		if err != nil {
			return errors.New("failed to create tail task")
		}
	}
	global.LOG.Info("task: %s", task.ID)

	// 把task的metadata都打印出来
	for k, v := range task.Metadata {
		global.LOG.Info("task metadata: %s=%v", k, v)
	}

	reader, err := global.LogStream.GetReader(task.ID)
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
		agentConn, err := conn.CENTER.GetAgentConn(&host)
		if err != nil {
			global.LOG.Error("get agent conn failed: %v", err)
			return fmt.Errorf("get agent conn failed: %w", err)
		}

		err = s.notifyRemote(agentConn, task.ID, task.LogPath, message.LogStreamStart, offset, whence)
		if err != nil {
			return fmt.Errorf("failed to start stream : %w", err)
		}
	}

	logCh, err := reader.Follow(offset, whence)
	if err != nil {
		global.LOG.Error("follow log failed: %v", err)
		return fmt.Errorf("follow log failed: %w", err)
	}
	global.LOG.Info("follow log success for task %s, path: %s", task.ID, path)

	// 获取任务状态监听器
	watcher, err := global.LogStream.GetTaskWatcher(task.ID)
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
	ctx := c.Request.Context()

	heartbeat := time.NewTicker(heartbeatInterval)
	defer heartbeat.Stop()

	// 创建一个缓冲通道来处理日志
	bufferCh := make(chan []byte, 100)
	defer close(bufferCh)

	// 更新一下任务状态
	// if task.Status == types.TaskStatusCreated {
	// 	if err := global.LogStream.UpdateTaskStatus(task.ID, types.TaskStatusRunning); err != nil {
	// 		global.LOG.Error("Failed to update task status to %s : %v", types.TaskStatusRunning, err)
	// 		return fmt.Errorf("Failed to update task status to %s : %w", types.TaskStatusRunning, err)
	// 	}
	// }

	// 设置 SSE 响应头
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

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

	defer func() {
		if r := recover(); r != nil {
			global.LOG.Warn("Recovered in SSE loop: %v", r)
		}
	}()

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
			global.LOG.Info("SSE DONE")
			// 如果是远程读取器，发送停止消息
			if _, ok := reader.(*adapters.RemoteReader); ok {
				// 获取agent连接
				agentConn, err := conn.CENTER.GetAgentConn(&host)
				if err != nil {
					global.LOG.Error("get agent conn failed: %v", err)
					return fmt.Errorf("get agent conn failed: %w", err)
				}

				go func() {
					err := s.notifyRemote(agentConn, task.ID, task.LogPath, message.LogStreamStop, 0, 0)
					if err != nil {
						global.LOG.Error("notify remote logstream message failed: %v", err)
					}
				}()
			}

			// 清理任务相关的资源
			// s.clearTaskStuff(task.ID)
			return nil
		}
	}
}

func (s *LogManService) notifyRemote(conn *net.Conn, taskId string, logPath string, msgType message.LogStreamType, offset int64, whence int) error {
	global.LOG.Info("notify remote logstream message %s", msgType)
	stopMsg, err := message.CreateLogStreamMessage(
		utils.GenerateMsgId(),
		msgType,
		taskId,
		logPath,
		offset,
		whence,
		"",
		"",
	)
	if err == nil {
		err = message.SendLogStreamMessage(*conn, stopMsg)
		if err != nil {
			global.LOG.Error("send logstream message failed: %v", err)
		}
	}
	return nil
}

// func (s *LogManService) clearTaskStuff(taskId string) {
// 	global.LOG.Info("clear task stuff")
// 	// 更新状态后删除task
// 	if err := global.LogStream.UpdateTaskStatus(taskId, types.TaskStatusCanceled); err != nil {
// 		global.LOG.Error("Failed to update task status to %s : %v", types.TaskStatusCanceled, err)
// 	}
// 	if err := global.LogStream.DeleteTask(taskId); err != nil {
// 		global.LOG.Error("delete task %s failed: %v", taskId, err)
// 	} else {
// 		global.LOG.Info("delete task %s success", taskId)
// 	}
// }

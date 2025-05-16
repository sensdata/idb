package docker

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/core/conn"
	"github.com/sensdata/idb/center/db/repo"
	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/logstream/pkg/reader/adapters"
	"github.com/sensdata/idb/core/logstream/pkg/types"
	"github.com/sensdata/idb/core/message"
	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/utils"
)

const (
	heartbeatInterval = 10 * time.Second // 心跳间隔
)

func (s *DockerMan) containerQuery(hostID uint64, req model.QueryContainer) (*model.PageResult, error) {
	var result model.PageResult

	data, err := utils.ToJSONString(req)
	if err != nil {
		return &result, err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Docker_Container_Query,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return &result, err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return &result, fmt.Errorf("failed to query container")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &result)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to container result: %v", err)
		return &result, fmt.Errorf("json err: %v", err)
	}

	return &result, nil
}

func (s *DockerMan) containerNames(hostID uint64) (*model.PageResult, error) {
	var result model.PageResult

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Docker_Container_Names,
			Data:   "",
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return &result, err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return &result, fmt.Errorf("failed to query container names")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &result)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to container names result: %v", err)
		return &result, fmt.Errorf("json err: %v", err)
	}

	return &result, nil
}

func (s *DockerMan) containerUsages(hostID uint64) (*model.PageResult, error) {
	var result model.PageResult

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Docker_Container_Resource_Usage,
			Data:   "",
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return &result, err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return &result, fmt.Errorf("failed to get container usages")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &result)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to container usages: %v", err)
		return &result, fmt.Errorf("json err: %v", err)
	}

	return &result, nil
}

func (s *DockerMan) containerLimit(hostID uint64) (*model.ContainerResourceLimit, error) {
	var result model.ContainerResourceLimit

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Docker_Container_Resource_Limit,
			Data:   "",
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return &result, err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return &result, fmt.Errorf("failed to get container limit")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &result)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to container limit: %v", err)
		return &result, fmt.Errorf("json err: %v", err)
	}

	return &result, nil
}

func (s *DockerMan) createContainer(hostID uint64, req model.ContainerOperate) error {
	global.LOG.Info("create container begin")
	data, err := utils.ToJSONString(req)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Docker_Container_Create,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return fmt.Errorf("failed to create container")
	}

	return nil
}

func (s *DockerMan) updateContainer(hostID uint64, req model.ContainerOperate) error {
	data, err := utils.ToJSONString(req)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Docker_Container_Update,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return fmt.Errorf("failed to update container")
	}

	return nil
}

func (s *DockerMan) upgradeContainer(hostID uint64, req model.ContainerUpgrade) error {
	data, err := utils.ToJSONString(req)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Docker_Container_Upgrade,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return fmt.Errorf("failed to upgrade container")
	}

	return nil
}

func (s *DockerMan) renameContainer(hostID uint64, req model.Rename) error {
	data, err := utils.ToJSONString(req)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Docker_Container_Rename,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return fmt.Errorf("failed to rename container")
	}

	return nil
}

func (s *DockerMan) operateContainer(hostID uint64, req model.ContainerOperation) error {
	data, err := utils.ToJSONString(req)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Docker_Container_Operation,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return fmt.Errorf("failed to operate container")
	}

	return nil
}

func (s *DockerMan) containerInfo(hostID uint64, containerID string) (*model.ContainerOperate, error) {
	var result model.ContainerOperate

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Docker_Container_Info,
			Data:   containerID,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return &result, err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return &result, fmt.Errorf("failed to get container info")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &result)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to container info: %v", err)
		return &result, fmt.Errorf("json err: %v", err)
	}

	return &result, nil
}

func (s *DockerMan) containerStats(hostID uint64, containerID string) (*model.ContainerStats, error) {
	var result model.ContainerStats

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Docker_Container_Stats,
			Data:   containerID,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return &result, err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return &result, fmt.Errorf("failed to get container stats")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &result)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to container stats: %v", err)
		return &result, fmt.Errorf("json err: %v", err)
	}

	return &result, nil
}

func (s *DockerMan) containerLogClean(hostID uint64, containerID string) error {
	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Docker_Container_Log_Clean,
			Data:   containerID,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return fmt.Errorf("failed to clean container log")
	}

	return nil
}

func (s *DockerMan) tailContainerLogs(hostID uint, containerID string, offset int64, whence int) (*model.FileContentPartRsp, error) {
	var fileContentPartRsp model.FileContentPartRsp

	req := model.FileContentPartReq{
		Path:   containerID,
		Lines:  offset,
		Whence: whence,
	}
	data, err := utils.ToJSONString(req)
	if err != nil {
		return &fileContentPartRsp, err
	}
	actionRequest := model.HostAction{
		HostID: hostID,
		Action: model.Action{
			Action: model.Docker_Container_Logs,
			Data:   data,
		},
	}
	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return &fileContentPartRsp, err
	}
	if !actionResponse.Data.Action.Result {
		global.LOG.Error("failed to get container logs part")
		return &fileContentPartRsp, fmt.Errorf("failed to get container logs part")
	}
	err = utils.FromJSONString(actionResponse.Data.Action.Data, &fileContentPartRsp)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to container logs part: %v", err)
		return &fileContentPartRsp, fmt.Errorf("json err: %v", err)
	}
	return &fileContentPartRsp, nil
}

func (s *DockerMan) followContainerLogs(c *gin.Context) error {
	defer func() {
		if r := recover(); r != nil {
			global.LOG.Error("Panic in tailContentStream: %v", r)
		}
	}()
	global.LOG.Info("tail start")

	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		global.LOG.Info("Invalid host")
		return errors.New("Invalid host")
	}

	containerID := c.Query("id")
	if containerID == "" {
		global.LOG.Info("Invalid container id")
		return errors.New("Invalid container id")
	}

	// 通过 containerInfo 获取容器详情
	containerInfo, err := s.containerInfo(hostID, containerID)
	if err != nil {
		global.LOG.Error("get container info failed: %v", err)
		return fmt.Errorf("get container info failed: %w", err)
	}

	// 判断 containerType
	containerType := "docker"
	configFilePaths := ""
	for _, label := range containerInfo.Labels {
		if label.Key == "com.docker.compose.project" {
			containerType = "compose"
			break
		}
	}

	// compose 场景下，需要处理 workingDir和config_files
	if containerType == "compose" {
		var workingDir, configFiles string
		for _, label := range containerInfo.Labels {
			if label.Key == "com.docker.compose.project.working_dir" {
				workingDir = label.Value
			}
			if label.Key == "com.docker.compose.project.config_files" {
				configFiles = label.Value
			}
		}
		if workingDir == "" || configFiles == "" {
			global.LOG.Error("workingDir or configFiles is empty")
			return errors.New("config files not found")
		}

		var result []string
		files := strings.Split(configFiles, ",")

		for _, file := range files {
			file = strings.TrimSpace(file)
			if file == "" {
				continue
			}
			if filepath.IsAbs(file) {
				// 已经是绝对路径，直接使用
				result = append(result, file)
			} else {
				// 相对路径，需要基于 workDir 拼接
				result = append(result, filepath.Join(workingDir, file))
			}
		}

		// 拼接成字符串(可能有多个配置文件)
		configFilePaths = strings.Join(result, ",")
	}

	// 构造容器参数 docker:containerID 或者 compose:config_files
	var containerParam string
	if containerType == "docker" {
		containerParam = fmt.Sprintf("%s:%s", containerType, containerID)
	} else {
		containerParam = fmt.Sprintf("%s:%s", containerType, configFilePaths)
	}

	// follow
	follow := ""
	f, _ := strconv.ParseBool(c.Query("follow"))
	if f {
		follow = "follow"
	}

	var offset int64
	tail, err := strconv.ParseUint(c.Query("tail"), 10, 32)
	if err != nil {
		offset = 0
	} else {
		offset = int64(tail)
	}

	var whence int
	since := c.Query("since")
	switch since {
	case "24h":
		whence = 24 * 60
	case "4h":
		whence = 4 * 60
	case "1h":
		whence = 60
	case "10m":
		whence = 10
	default:
		whence = 0
	}

	// 找host
	hostRepo := repo.NewHostRepo()
	host, err := hostRepo.Get(hostRepo.WithByID(uint(hostID)))
	if err != nil {
		global.LOG.Error("get host failed: %v", err)
		return fmt.Errorf("get host failed: %w", err)
	}

	// 创建任务
	metadata := map[string]interface{}{
		"log_path": containerParam, // 容器参数 containerID 或者 containerID:config_files
	}
	task, err := global.LogStream.CreateTask(types.TaskTypeRemote, metadata)
	if err != nil {
		global.LOG.Info("failed to create task")
		return errors.New("failed to create tail task")
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

		err = s.notifyRemote(agentConn, task.ID, task.LogPath, message.LogStreamStart, offset, whence, follow)
		if err != nil {
			return fmt.Errorf("failed to start stream : %w", err)
		}
	}

	logCh, err := reader.Follow(offset, whence)
	if err != nil {
		global.LOG.Error("follow log failed: %v", err)
		return fmt.Errorf("follow log failed: %w", err)
	}
	global.LOG.Info("follow log success for task %s, path: %s", task.ID, containerParam)

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

				go s.notifyRemote(agentConn, task.ID, task.LogPath, message.LogStreamStop, 0, 0, "")
			}
			// 清理任务相关的资源
			//s.clearTaskStuff(task.ID)
			return nil
		}
	}
}

func (s *DockerMan) notifyRemote(conn *net.Conn, taskId string, logPath string, msgType message.LogStreamType, offset int64, whence int, content string) error {
	global.LOG.Info("notify remote logstream message %s", msgType)
	stopMsg, err := message.CreateLogStreamMessage(
		utils.GenerateMsgId(),
		msgType,
		taskId,
		logPath,
		offset,
		whence,
		content,
		"",
	)
	if err == nil {
		message.SendLogStreamMessage(*conn, stopMsg)
	}
	return nil
}

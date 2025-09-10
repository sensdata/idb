package docker

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/sensdata/idb/center/core/conn"
	"github.com/sensdata/idb/center/db/repo"
	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/logstream/pkg/reader/adapters"
	"github.com/sensdata/idb/core/logstream/pkg/types"
	"github.com/sensdata/idb/core/message"
	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/utils"
)

func (s *DockerMan) composeQuery(hostID uint64, req model.QueryCompose) (*model.PageResult, error) {
	var result model.PageResult

	data, err := utils.ToJSONString(req)
	if err != nil {
		return &result, err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Docker_Compose_Page,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return &result, err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return &result, errors.New(actionResponse.Data.Action.Data)
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &result)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to compose result: %v", err)
		return &result, fmt.Errorf("json err: %v", err)
	}

	return &result, nil
}

func (s *DockerMan) composeCreate(hostID uint64, req model.CreateCompose) (*model.ComposeCreateResult, error) {
	var result model.ComposeCreateResult

	composeCreate := model.ComposeCreate{
		Name:           req.Name,
		ComposeContent: req.ComposeContent,
		EnvContent:     req.EnvContent,
		ConfPath:       "",
		ConfContent:    "",
		WorkDir:        s.AppDir,
	}
	data, err := utils.ToJSONString(composeCreate)
	if err != nil {
		return &result, err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Docker_Compose_Create,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return &result, err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return &result, fmt.Errorf("failed to create compose")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &result)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to compose create result: %v", err)
		return &result, fmt.Errorf("json err: %v", err)
	}

	return &result, nil
}

func (s *DockerMan) composeDetail(hostID uint64, req model.ComposeDetailReq) (*model.ComposeDetailRsp, error) {
	var result model.ComposeDetailRsp
	data, err := utils.ToJSONString(req)
	if err != nil {
		return &result, err
	}
	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Docker_Compose_Detail,
			Data:   data,
		},
	}
	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return &result, err
	}
	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action Docker_Compose_Detail failed")
		return &result, fmt.Errorf("failed to query compose detail")
	}
	err = utils.FromJSONString(actionResponse.Data.Action.Data, &result)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to compose detail result: %v", err)
		return &result, fmt.Errorf("json err: %v", err)
	}
	return &result, nil
}

func (s *DockerMan) composeUpdate(hostID uint64, req model.CreateCompose) (*model.ComposeCreateResult, error) {
	var result model.ComposeCreateResult
	composeUpdate := model.ComposeUpdate{
		Name:           req.Name,
		ComposeContent: req.ComposeContent,
		EnvContent:     req.EnvContent,
		WorkDir:        s.AppDir,
	}

	data, err := utils.ToJSONString(composeUpdate)
	if err != nil {
		return &result, err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Docker_Compose_Update,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return &result, err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return &result, fmt.Errorf("failed to update compose")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &result)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to compose update result: %v", err)
		return &result, fmt.Errorf("json err: %v", err)
	}

	return &result, nil
}

func (s *DockerMan) composeRemove(hostID uint64, req model.ComposeRemove) (*model.ComposeCreateResult, error) {
	var result model.ComposeCreateResult
	data, err := utils.ToJSONString(req)
	if err != nil {
		return &result, err
	}
	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Docker_Compose_Remove,
			Data:   data,
		},
	}
	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return &result, err
	}
	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action Docker_Compose_Remove failed")
		return &result, fmt.Errorf("failed to remove compose: %s", actionResponse.Data)
	}
	err = utils.FromJSONString(actionResponse.Data.Action.Data, &result)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to compose remove result: %v", err)
		return &result, fmt.Errorf("json err: %v", err)
	}
	return &result, nil
}

func (s *DockerMan) composeTest(hostID uint64, req model.CreateCompose) (*model.ComposeTestResult, error) {
	var result model.ComposeTestResult

	composeCreate := model.ComposeCreate{
		Name:           req.Name,
		ComposeContent: req.ComposeContent,
		EnvContent:     req.EnvContent,
		ConfPath:       "",
		ConfContent:    "",
		WorkDir:        s.AppDir,
	}
	data, err := utils.ToJSONString(composeCreate)
	if err != nil {
		return &result, err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Docker_Compose_Test,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return &result, err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return &result, fmt.Errorf("failed to test compose")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &result)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to compose test result: %v", err)
		return &result, fmt.Errorf("json err: %v", err)
	}

	return &result, nil
}

func (s *DockerMan) composeOperation(hostID uint64, req model.OperateCompose) (*model.OperationResult, error) {
	var result model.OperationResult = model.OperationResult{
		Success: false,
		Message: constant.OperationFailed,
		Command: fmt.Sprintf("docker compose %s", req.Operation),
	}

	composeOperation := model.ComposeOperation{
		Name:          req.Name,
		Operation:     req.Operation,
		RemoveVolumes: req.RemoveVolumes,
		WorkDir:       s.AppDir,
	}
	data, err := utils.ToJSONString(composeOperation)
	if err != nil {
		return &result, err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Docker_Compose_Operation,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return &result, err
	}

	result.Success = actionResponse.Data.Action.Result
	if !result.Success {
		global.LOG.Error("compose operation failed")
		return &result, fmt.Errorf("failed to operate compose")
	} else {
		var composeCreateResult model.ComposeCreateResult
		err = utils.FromJSONString(actionResponse.Data.Action.Data, &composeCreateResult)
		if err != nil {
			global.LOG.Error("Error unmarshaling data to compose create result: %v", err)
			return &result, fmt.Errorf("json err: %v", err)
		}

		result.Message = constant.OperationSuccess
		result.Extra = composeCreateResult.Log
	}

	return &result, nil
}

func (s *DockerMan) followComposeLogs(c *gin.Context) error {
	defer func() {
		if r := recover(); r != nil {
			global.LOG.Error("Panic in followComposeLogs: %v", r)
		}
	}()
	global.LOG.Info("tail start")

	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		global.LOG.Info("Invalid host")
		return errors.New("Invalid host")
	}

	config_files := c.Query("config_files")
	if config_files == "" {
		global.LOG.Info("Invalid name")
		return errors.New("Invalid name")
	}

	follow := "follow"
	offset := 0
	whence := 0

	// 找host
	hostRepo := repo.NewHostRepo()
	host, err := hostRepo.Get(hostRepo.WithByID(uint(hostID)))
	if err != nil {
		global.LOG.Error("get host failed: %v", err)
		return fmt.Errorf("get host failed: %w", err)
	}

	// 创建任务
	composeParam := fmt.Sprintf("compose:%s", config_files)
	metadata := map[string]interface{}{
		"log_path": composeParam, // compose:composePath
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

		err = s.notifyRemote(agentConn, task.ID, task.LogPath, message.LogStreamStart, int64(offset), whence, follow)
		if err != nil {
			return fmt.Errorf("failed to start stream : %w", err)
		}
	}

	logCh, err := reader.Follow(int64(offset), whence)
	if err != nil {
		global.LOG.Error("follow log failed: %v", err)
		return fmt.Errorf("follow log failed: %w", err)
	}
	global.LOG.Info("follow log success for task %s, path: %s", task.ID, composeParam)

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

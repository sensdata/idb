package fileman

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
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

const (
	heartbeatInterval = 10 * time.Second // 心跳间隔
)

func (s *FileMan) sendAction(actionRequest model.HostAction) (*model.ActionResponse, error) {
	var actionResponse model.ActionResponse

	resp, err := s.restyClient.R().
		SetBody(actionRequest).
		SetResult(&actionResponse).
		Post("/actions") // 修改URL路径

	if err != nil {
		global.LOG.Error("failed to send request: %v", err)
		return nil, fmt.Errorf("failed to send request: %v", err)
	}

	if resp.StatusCode() != 200 {
		global.LOG.Error("received error response: %s", resp.Status())
		return nil, fmt.Errorf("received error response: %s", resp.Status())
	}

	return &actionResponse, nil
}

func (s *FileMan) getFileTree(hostID uint64, op model.FileOption) ([]model.FileTree, error) {
	var fileTree []model.FileTree

	data, err := utils.ToJSONString(op)
	if err != nil {
		return fileTree, err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.File_Tree,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return fileTree, err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return fileTree, fmt.Errorf("failed to get filetree")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &fileTree)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to filetree: %v", err)
		return fileTree, fmt.Errorf("json err: %v", err)
	}

	return fileTree, nil
}

func (s *FileMan) getFileList(hostID uint64, op model.FileOption) (*model.FileInfo, error) {
	var fileInfo model.FileInfo
	data, err := utils.ToJSONString(op)
	if err != nil {
		return &fileInfo, err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.File_List,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return &fileInfo, err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return &fileInfo, fmt.Errorf("failed to get file list")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &fileInfo)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to file list: %v", err)
		return &fileInfo, fmt.Errorf("json err: %v", err)
	}

	return &fileInfo, nil
}

func (s *FileMan) searchFile(hostID uint64, op model.FileOption) (*model.PageResult, error) {
	var result model.PageResult
	data, err := utils.ToJSONString(op)
	if err != nil {
		return &result, err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.File_Search,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return &result, err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return &result, fmt.Errorf("failed to search file")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &result)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to file list: %v", err)
		return &result, fmt.Errorf("json err: %v", err)
	}

	return &result, nil
}

func (s *FileMan) create(hostID uint64, op model.FileCreate) error {
	data, err := utils.ToJSONString(op)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.File_Create,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return errors.New(actionResponse.Data.Action.Data)
	}

	return nil
}
func (s *FileMan) delete(hostID uint64, op model.FileDelete) error {
	data, err := utils.ToJSONString(op)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.File_Delete,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return fmt.Errorf("failed to delete")
	}

	return nil
}
func (s *FileMan) batchDelete(hostID uint64, op model.FileBatchDelete) error {
	data, err := utils.ToJSONString(op)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.File_Batch_Delete,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return fmt.Errorf("failed to batch delete")
	}

	return nil
}
func (s *FileMan) compress(hostID uint64, op model.FileCompress) error {
	data, err := utils.ToJSONString(op)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.File_Compress,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return fmt.Errorf("failed to compress")
	}

	return nil
}
func (s *FileMan) decompress(hostID uint64, op model.FileDeCompress) error {
	data, err := utils.ToJSONString(op)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.File_Decompress,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return errors.New(actionResponse.Data.Action.Data)
	}

	return nil
}
func (s *FileMan) getContent(hostID uint64, op model.FileContentReq) (*model.FileInfo, error) {
	var fileInfo model.FileInfo
	data, err := utils.ToJSONString(op)
	if err != nil {
		return &fileInfo, err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.File_Content,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return &fileInfo, err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		if strings.Contains(actionResponse.Data.Action.Data, "no such file or directory") {
			return &fileInfo, constant.ErrFileNotExist
		}
		return &fileInfo, fmt.Errorf("failed to get file content")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &fileInfo)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to file content: %v", err)
		return &fileInfo, fmt.Errorf("json err: %v", err)
	}

	return &fileInfo, nil
}

func (s *FileMan) saveContent(hostID uint64, op model.FileEdit) error {
	data, err := utils.ToJSONString(op)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.File_Content_Modify,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return fmt.Errorf("failed to edit content")
	}

	return nil
}

func (s *FileMan) getContentPart(hostID uint, path string, lines int64, whence int) (*model.FileContentPartRsp, error) {
	var fileContentPartRsp model.FileContentPartRsp

	req := model.FileContentPartReq{
		Path:   path,
		Lines:  lines,
		Whence: whence,
	}
	data, err := utils.ToJSONString(req)
	if err != nil {
		return &fileContentPartRsp, err
	}
	actionRequest := model.HostAction{
		HostID: hostID,
		Action: model.Action{
			Action: model.File_Content_Part,
			Data:   data,
		},
	}
	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return &fileContentPartRsp, err
	}
	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return &fileContentPartRsp, fmt.Errorf("failed to get file content part")
	}
	err = utils.FromJSONString(actionResponse.Data.Action.Data, &fileContentPartRsp)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to file content part: %v", err)
		return &fileContentPartRsp, fmt.Errorf("json err: %v", err)
	}
	return &fileContentPartRsp, nil
}

func (s *FileMan) tailContentStream(c *gin.Context) error {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		return errors.New("invalid host")
	}

	path := c.Query("path")
	if path == "" {
		return errors.New("invalid path")
	}

	var offset int64 = 0
	var whence int = io.SeekEnd

	// 找host
	hostRepo := repo.NewHostRepo()
	host, err := hostRepo.Get(hostRepo.WithByID(uint(hostID)))
	if err != nil {
		global.LOG.Error("get host failed: %v", err)
		return fmt.Errorf("get host failed: %w", err)
	}

	// 查找任务
	var task *types.Task
	task, err = global.LogStream.GetTaskByLog(path)
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
	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	heartbeat := time.NewTicker(heartbeatInterval)
	defer heartbeat.Stop()

	// 创建一个缓冲通道来处理日志
	bufferCh := make(chan []byte, 100)
	defer close(bufferCh)

	// 更新一下任务状态
	if task.Status == types.TaskStatusCreated {
		if err := global.LogStream.UpdateTaskStatus(task.ID, types.TaskStatusRunning); err != nil {
			global.LOG.Error("Failed to update task status to %s : %v", types.TaskStatusRunning, err)
			return fmt.Errorf("Failed to update task status to %s : %w", types.TaskStatusRunning, err)
		}
	}

	// 设置 SSE 响应头
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")

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

	notify := c.Writer.CloseNotify()
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
		case <-notify:
			global.LOG.Info("SSE NOTIFY")
			// 如果是远程读取器，发送停止消息
			if _, ok := reader.(*adapters.RemoteReader); ok {
				// 获取agent连接
				agentConn, err := conn.CENTER.GetAgentConn(&host)
				if err != nil {
					global.LOG.Error("get agent conn failed: %v", err)
					return fmt.Errorf("get agent conn failed: %w", err)
				}

				go s.notifyRemote(agentConn, task.ID, task.LogPath, message.LogStreamStop, 0, 0)
			}
			// 清理任务相关的资源
			go s.clearTaskStuff(task.ID)
			return nil
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

				go s.notifyRemote(agentConn, task.ID, task.LogPath, message.LogStreamStop, 0, 0)
			}
			// 清理任务相关的资源
			go s.clearTaskStuff(task.ID)
			return nil
		}
	}
}

func (s *FileMan) notifyRemote(conn *net.Conn, taskId string, logPath string, msgType message.LogStreamType, offset int64, whence int) error {
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
		message.SendLogStreamMessage(*conn, stopMsg)
	}
	return nil
}

func (s *FileMan) clearTaskStuff(taskId string) {
	global.LOG.Info("clear task stuff")
	// 更新状态后删除task
	if err := global.LogStream.UpdateTaskStatus(taskId, types.TaskStatusCanceled); err != nil {
		global.LOG.Error("Failed to update task status to %s : %v", types.TaskStatusCanceled, err)
	}
	if err := global.LogStream.DeleteTask(taskId); err != nil {
		global.LOG.Error("delete task %s failed: %v", taskId, err)
	} else {
		global.LOG.Info("delete task %s success", taskId)
	}
}

func (s *FileMan) uploadFile(hostID uint, path string, file *multipart.FileHeader) error {
	return conn.CENTER.UploadFile(hostID, path, file)
}

func (s *FileMan) downloadFile(c *gin.Context, hostID uint, path string) error {
	return conn.CENTER.DownloadFile(c, hostID, path)
}

func (s *FileMan) dirSize(hostID uint64, op model.DirSizeReq) (*model.DirSizeRes, error) {
	var dirSize model.DirSizeRes
	data, err := utils.ToJSONString(op)
	if err != nil {
		return &dirSize, err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.File_Dir_Size,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return &dirSize, err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return &dirSize, fmt.Errorf("failed to get file content")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &dirSize)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to file content: %v", err)
		return &dirSize, fmt.Errorf("json err: %v", err)
	}

	return &dirSize, nil
}
func (s *FileMan) changeName(hostID uint64, op model.FileRename) error {
	data, err := utils.ToJSONString(op)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.File_Change_Name,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return fmt.Errorf("failed to change name")
	}

	return nil
}
func (s *FileMan) mvFile(hostID uint64, op model.FileMove) error {
	data, err := utils.ToJSONString(op)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.File_Move,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return fmt.Errorf("failed to mv file")
	}

	return nil
}
func (s *FileMan) changeOwner(hostID uint64, op model.FileRoleUpdate) error {
	data, err := utils.ToJSONString(op)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.File_Change_Owner,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return fmt.Errorf("failed to change owner")
	}

	return nil
}
func (s *FileMan) changeMode(hostID uint64, op model.FileCreate) error {
	data, err := utils.ToJSONString(op)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.File_Change_Mode,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return fmt.Errorf("failed to change mode")
	}

	return nil
}

func (s *FileMan) batchChangeMode(hostID uint64, op model.FileModeReq) error {
	data, err := utils.ToJSONString(op)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.File_Batch_Change_Mode,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return fmt.Errorf("failed to batch change mode")
	}

	return nil
}

func (s *FileMan) batchChangeOwner(hostID uint64, op model.FileRoleReq) error {
	data, err := utils.ToJSONString(op)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.File_Batch_Change_Owner,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return fmt.Errorf("failed to batch change owner")
	}

	return nil
}

func (s *FileMan) getFavoriteList(hostID uint64, req model.FavoriteListReq) (*model.PageResult, error) {
	var pageResult model.PageResult
	data, err := utils.ToJSONString(req.PageInfo)
	if err != nil {
		return &pageResult, err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Favorite_List,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return &pageResult, err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return &pageResult, fmt.Errorf("failed to get fav list")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &pageResult)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to fav list: %v", err)
		return &pageResult, fmt.Errorf("json err: %v", err)
	}

	return &pageResult, nil
}
func (s *FileMan) createFavorite(hostID uint64, req model.FavoriteCreate) (*model.Favorite, error) {
	var favorite model.Favorite
	data, err := utils.ToJSONString(req)
	if err != nil {
		return &favorite, err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Favorite_Create,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return &favorite, err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return &favorite, fmt.Errorf("failed to create fav")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &favorite)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to file list: %v", err)
		return &favorite, fmt.Errorf("json err: %v", err)
	}

	return &favorite, nil
}

func (s *FileMan) deleteFavorite(hostID uint64, req model.FavoriteDelete) error {
	data, err := utils.ToJSONString(req)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Favorite_Delete,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return fmt.Errorf("failed to delete fav")
	}

	return nil
}

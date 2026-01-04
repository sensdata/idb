package service

import (
	"fmt"

	"github.com/sensdata/idb/center/core/conn"
	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/utils"
)

type RsyncService struct{}

type IRSyncService interface {
	ListTask(hostID uint, req model.RsyncListTaskRequest) (*model.RsyncClientListTaskResponse, error)
	QueryTask(hostID uint, req model.RsyncQueryTaskRequest) (*model.RsyncClientTask, error)
	CreateTask(hostID uint, req model.RsyncClientCreateTaskRequest) (*model.RsyncCreateTaskResponse, error)
	UpdateTask(hostID uint, req model.RsyncClientUpdateTaskRequest) error
	DeleteTask(hostID uint, req model.RsyncDeleteTaskRequest) error
	CancelTask(hostID uint, req model.RsyncCancelTaskRequest) error
	RetryTask(hostID uint, req model.RsyncRetryTaskRequest) error
	TestTask(hostID uint, req model.RsyncTestTaskRequest) (*model.RsyncTaskLog, error)
	TaskLogList(hostID uint, req model.RsyncTaskLogListRequest) (*model.RsyncTaskLogListResponse, error)
}

func NewIRsyncService() IRSyncService {
	return &RsyncService{}
}

func (s *RsyncService) ListTask(hostID uint, req model.RsyncListTaskRequest) (*model.RsyncClientListTaskResponse, error) {
	var resp model.RsyncClientListTaskResponse

	data, err := utils.ToJSONString(req)
	if err != nil {
		return &resp, err
	}

	actionRequest := model.HostAction{
		HostID: hostID,
		Action: model.Action{
			Action: model.Rsync_List,
			Data:   data,
		},
	}

	actionResponse, err := conn.CENTER.ExecuteAction(actionRequest)
	if err != nil {
		return &resp, err
	}
	if !actionResponse.Result {
		global.LOG.Error("action Rsync_List failed")
		return &resp, fmt.Errorf("failed to list rsync tasks: %s", actionResponse.Data)
	}

	err = utils.FromJSONString(actionResponse.Data, &resp)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to rsync tasks: %v", err)
		return &resp, fmt.Errorf("json err: %v", err)
	}

	return &resp, nil
}

func (s *RsyncService) QueryTask(hostID uint, req model.RsyncQueryTaskRequest) (*model.RsyncClientTask, error) {
	var resp model.RsyncClientTask

	data, err := utils.ToJSONString(req)
	if err != nil {
		return &resp, err
	}

	actionRequest := model.HostAction{
		HostID: hostID,
		Action: model.Action{
			Action: model.Rsync_Detail,
			Data:   data,
		},
	}

	actionResponse, err := conn.CENTER.ExecuteAction(actionRequest)
	if err != nil {
		return &resp, err
	}
	if !actionResponse.Result {
		global.LOG.Error("action Rsync_Detail failed")
		return &resp, fmt.Errorf("failed to query rsync task: %s", actionResponse.Data)
	}

	err = utils.FromJSONString(actionResponse.Data, &resp)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to rsync task: %v", err)
		return &resp, fmt.Errorf("json err: %v", err)
	}

	return &resp, nil
}

func (s *RsyncService) CreateTask(hostID uint, req model.RsyncClientCreateTaskRequest) (*model.RsyncCreateTaskResponse, error) {
	var resp model.RsyncCreateTaskResponse

	data, err := utils.ToJSONString(req)
	if err != nil {
		return &resp, err
	}

	actionRequest := model.HostAction{
		HostID: hostID,
		Action: model.Action{
			Action: model.Rsync_Create,
			Data:   data,
		},
	}

	actionResponse, err := conn.CENTER.ExecuteAction(actionRequest)
	if err != nil {
		return &resp, err
	}
	if !actionResponse.Result {
		global.LOG.Error("action Rsync_Create failed")
		return &resp, fmt.Errorf("failed to create rsync task: %s", actionResponse.Data)
	}

	err = utils.FromJSONString(actionResponse.Data, &resp)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to rsync create task response: %v", err)
		return &resp, fmt.Errorf("json err: %v", err)
	}

	return &resp, nil
}

func (s *RsyncService) UpdateTask(hostID uint, req model.RsyncClientUpdateTaskRequest) error {
	data, err := utils.ToJSONString(req)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: hostID,
		Action: model.Action{
			Action: model.Rsync_Update,
			Data:   data,
		},
	}

	actionResponse, err := conn.CENTER.ExecuteAction(actionRequest)
	if err != nil {
		return err
	}
	if !actionResponse.Result {
		global.LOG.Error("action Rsync_Update failed")
		return fmt.Errorf("failed to update rsync task: %s", actionResponse.Data)
	}

	return nil
}

func (s *RsyncService) DeleteTask(hostID uint, req model.RsyncDeleteTaskRequest) error {
	data, err := utils.ToJSONString(req)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: hostID,
		Action: model.Action{
			Action: model.Rsync_Delete,
			Data:   data,
		},
	}

	actionResponse, err := conn.CENTER.ExecuteAction(actionRequest)
	if err != nil {
		return err
	}
	if !actionResponse.Result {
		global.LOG.Error("action Rsync_Delete failed")
		return fmt.Errorf("failed to delete rsync task: %s", actionResponse.Data)
	}

	return nil
}

func (s *RsyncService) CancelTask(hostID uint, req model.RsyncCancelTaskRequest) error {
	data, err := utils.ToJSONString(req)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: hostID,
		Action: model.Action{
			Action: model.Rsync_Stop,
			Data:   data,
		},
	}

	actionResponse, err := conn.CENTER.ExecuteAction(actionRequest)
	if err != nil {
		return err
	}
	if !actionResponse.Result {
		global.LOG.Error("action Rsync_Stop failed")
		return fmt.Errorf("failed to stop rsync task: %s", actionResponse.Data)
	}

	return nil
}

func (s *RsyncService) RetryTask(hostID uint, req model.RsyncRetryTaskRequest) error {
	data, err := utils.ToJSONString(req)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: hostID,
		Action: model.Action{
			Action: model.Rsync_Retry,
			Data:   data,
		},
	}

	actionResponse, err := conn.CENTER.ExecuteAction(actionRequest)
	if err != nil {
		return err
	}
	if !actionResponse.Result {
		global.LOG.Error("action Rsync_Retry failed")
		return fmt.Errorf("failed to retry rsync task: %s", actionResponse.Data)
	}

	return nil
}

func (s *RsyncService) TestTask(hostID uint, req model.RsyncTestTaskRequest) (*model.RsyncTaskLog, error) {
	var resp model.RsyncTaskLog

	data, err := utils.ToJSONString(req)
	if err != nil {
		return &resp, err
	}

	actionRequest := model.HostAction{
		HostID: hostID,
		Action: model.Action{
			Action: model.Rsync_Test,
			Data:   data,
		},
	}

	actionResponse, err := conn.CENTER.ExecuteAction(actionRequest)
	if err != nil {
		return &resp, err
	}
	if !actionResponse.Result {
		global.LOG.Error("action Rsync_Test failed")
		return &resp, fmt.Errorf("failed to test rsync task: %s", actionResponse.Data)
	}

	err = utils.FromJSONString(actionResponse.Data, &resp)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to rsync test task response: %v", err)
		return &resp, fmt.Errorf("json err: %v", err)
	}

	return &resp, nil
}

func (s *RsyncService) TaskLogList(hostID uint, req model.RsyncTaskLogListRequest) (*model.RsyncTaskLogListResponse, error) {
	var resp *model.RsyncTaskLogListResponse

	data, err := utils.ToJSONString(req)
	if err != nil {
		return resp, err
	}

	actionRequest := model.HostAction{
		HostID: hostID,
		Action: model.Action{
			Action: model.Rsync_Logs,
			Data:   data,
		},
	}

	actionResponse, err := conn.CENTER.ExecuteAction(actionRequest)
	if err != nil {
		return resp, err
	}
	if !actionResponse.Result {
		global.LOG.Error("action Rsync_Logs failed")
		return resp, fmt.Errorf("failed to get rsync task logs: %s", actionResponse.Data)
	}

	err = utils.FromJSONString(actionResponse.Data, &resp)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to rsync logs response: %v", err)
		return resp, fmt.Errorf("json err: %v", err)
	}

	return resp, nil
}

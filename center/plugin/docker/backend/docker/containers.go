package docker

import (
	"fmt"

	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/utils"
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

func (s *DockerMan) deleteContainer(hostID uint64, containerID string) error {
	return nil
}

// func (s *DockerMan) getContainerLog(hostID uint64, containerID string) (*model.ContainerLog, error) {
// 	return nil, nil
// }

func (s *DockerMan) getContainerStatus(hostID uint64, containerID string) (*model.ContainerStats, error) {
	return nil, nil
}

func (s *DockerMan) startContainer(hostID uint64, containerID string) error {
	return nil
}

func (s *DockerMan) stopContainer(hostID uint64, containerID string, force bool) error {
	return nil
}

func (s *DockerMan) rebootContainer(hostID uint64, containerID string) error {
	return nil
}

func (s *DockerMan) pauseContainer(hostID uint64, containerID string) error {
	return nil
}

func (s *DockerMan) resumeContainer(hostID uint64, containerID string) error {
	return nil
}

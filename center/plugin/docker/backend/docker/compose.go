package docker

import (
	"fmt"

	"github.com/sensdata/idb/center/global"
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
		return &result, fmt.Errorf("failed to query compose")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &result)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to compose result: %v", err)
		return &result, fmt.Errorf("json err: %v", err)
	}

	return &result, nil
}

func (s *DockerMan) composeCreate(hostID uint64, req model.ComposeCreate) (*model.ComposeCreateResult, error) {
	var result model.ComposeCreateResult

	data, err := utils.ToJSONString(req)
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

func (s *DockerMan) composeUpdate(hostID uint64, req model.ComposeUpdate) error {
	data, err := utils.ToJSONString(req)
	if err != nil {
		return err
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
		return err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return fmt.Errorf("failed to update compose")
	}

	return nil
}

func (s *DockerMan) composeTest(hostID uint64, req model.ComposeCreate) (*model.ComposeTestResult, error) {
	var result model.ComposeTestResult

	data, err := utils.ToJSONString(req)
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

func (s *DockerMan) composeOperation(hostID uint64, req model.ComposeOperation) error {
	data, err := utils.ToJSONString(req)
	if err != nil {
		return err
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
		return err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return fmt.Errorf("failed to operate compose")
	}

	return nil
}

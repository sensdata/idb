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

func (s *DockerMan) composeUpdate(hostID uint64, req model.CreateCompose) error {
	composeUpdate := model.ComposeUpdate{
		Name:           req.Name,
		ComposeContent: req.ComposeContent,
		EnvContent:     req.EnvContent,
		WorkDir:        s.AppDir,
	}

	data, err := utils.ToJSONString(composeUpdate)
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

func (s *DockerMan) composeRemove(hostID uint64, req model.ComposeRemove) error {
	data, err := utils.ToJSONString(req)
	if err != nil {
		return err
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
		return err
	}
	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action Docker_Compose_Remove failed")
		return fmt.Errorf("failed to remove compose: %s", actionResponse.Data)
	}
	return nil
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

func (s *DockerMan) composeOperation(hostID uint64, req model.OperateCompose) error {

	composeOperation := model.ComposeOperation{
		Name:      req.Name,
		Operation: req.Operation,
		WorkDir:   s.AppDir,
	}
	data, err := utils.ToJSONString(composeOperation)
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

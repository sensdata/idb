package docker

import (
	"fmt"

	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/utils"
)

func (s *DockerMan) getNetworks(hostID uint64, req model.SearchPageInfo) (*model.PageResult, error) {
	var result model.PageResult

	data, err := utils.ToJSONString(req)
	if err != nil {
		return &result, err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Docker_Network_Page,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return &result, err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return &result, fmt.Errorf("failed to query networks")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &result)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to networks result: %v", err)
		return &result, fmt.Errorf("json err: %v", err)
	}

	return &result, nil
}

func (s *DockerMan) networkNames(hostID uint64) (*model.PageResult, error) {
	var result model.PageResult

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Docker_Network_List,
			Data:   "",
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return &result, err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return &result, fmt.Errorf("failed to query network names")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &result)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to network names result: %v", err)
		return &result, fmt.Errorf("json err: %v", err)
	}

	return &result, nil
}

func (s *DockerMan) deleteNetwork(hostID uint64, req model.BatchDelete) error {
	data, err := utils.ToJSONString(req)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Docker_Network_Delete,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return fmt.Errorf("failed to remove networks")
	}

	return nil
}

func (s *DockerMan) createNetwork(hostID uint64, req model.NetworkCreate) error {
	data, err := utils.ToJSONString(req)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Docker_Network_Create,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return fmt.Errorf("failed to create network")
	}

	return nil
}

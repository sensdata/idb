package process

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/utils"
)

func (s *Process) getProcessList(hostID uint, req model.ProcessListRequest) (*model.ProcessListResponse, error) {
	var processListResponse model.ProcessListResponse

	data, err := utils.ToJSONString(req)
	if err != nil {
		return &processListResponse, err
	}

	actionRequest := model.HostAction{
		HostID: hostID,
		Action: model.Action{
			Action: model.PS_List,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return &processListResponse, err
	}

	if !actionResponse.Data.Action.Result {
		LOG.Error("action PS_List failed: %s", actionResponse.Data.Action.Data)
		return &processListResponse, errors.New(actionResponse.Data.Action.Data)
	}

	err = json.Unmarshal([]byte(actionResponse.Data.Action.Data), &processListResponse)
	if err != nil {
		LOG.Error("Error unmarshaling data to process list: %v", err)
	}

	return &processListResponse, nil
}

func (s *Process) getProcessDetail(hostID uint, req model.ProcessRequest) (*model.ProcessDetail, error) {
	var processDetailResponse model.ProcessDetail

	data, err := utils.ToJSONString(req)
	if err != nil {
		return &processDetailResponse, err
	}

	actionRequest := model.HostAction{
		HostID: hostID,
		Action: model.Action{
			Action: model.PS_Detail,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return &processDetailResponse, err
	}

	if !actionResponse.Data.Action.Result {
		LOG.Error("action PS_Detail failed: %s", actionResponse.Data.Action.Data)
		return &processDetailResponse, errors.New(actionResponse.Data.Action.Data)
	}

	err = json.Unmarshal([]byte(actionResponse.Data.Action.Data), &processDetailResponse)
	if err != nil {
		LOG.Error("Error unmarshaling data to process detail: %v", err)
		return &processDetailResponse, fmt.Errorf("json err: %v", err)
	}

	return &processDetailResponse, nil
}

func (s *Process) killProcess(hostID uint, req model.ProcessRequest) error {
	data, err := utils.ToJSONString(req)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: hostID,
		Action: model.Action{
			Action: model.PS_Kill,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		LOG.Error("action PS_Kill failed: %s", actionResponse.Data.Action.Data)
		return errors.New(actionResponse.Data.Action.Data)
	}

	return nil
}

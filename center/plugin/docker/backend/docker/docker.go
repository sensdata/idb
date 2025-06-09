package docker

import (
	"fmt"

	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/utils"
)

func (s *DockerMan) dockerStatus(hostID uint64) (*model.DockerStatus, error) {
	var status model.DockerStatus

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Docker_Status,
			Data:   "",
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return &status, err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return &status, fmt.Errorf("failed to get docker status")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &status)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to docker status: %v", err)
		return &status, fmt.Errorf("json err: %v", err)
	}

	return &status, nil
}

func (s *DockerMan) dockerConf(hostID uint64) (*model.DaemonJsonConf, error) {
	var conf model.DaemonJsonConf

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Docker_Conf,
			Data:   "",
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return &conf, err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return &conf, fmt.Errorf("failed to get docker conf")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &conf)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to docker conf: %v", err)
		return &conf, fmt.Errorf("json err: %v", err)
	}

	return &conf, nil
}

func (s *DockerMan) dockerUpdateConf(hostID uint64, req model.KeyValue) error {

	data, err := utils.ToJSONString(req)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Docker_Upd_Conf,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return fmt.Errorf("failed to update docker conf")
	}

	return nil
}

func (s *DockerMan) dockerUpdateConfByFile(hostID uint64, req model.DaemonJsonUpdateRaw) error {

	data, err := utils.ToJSONString(req)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Docker_Upd_Conf_File,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return fmt.Errorf("failed to update docker conf by file")
	}

	return nil
}

func (s *DockerMan) dockerUpdateLogOption(hostID uint64, req model.LogOption) error {

	data, err := utils.ToJSONString(req)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Docker_Upd_Log,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return fmt.Errorf("failed to update log option")
	}

	return nil
}

func (s *DockerMan) dockerUpdateIpv6Option(hostID uint64, req model.Ipv6Option) error {

	data, err := utils.ToJSONString(req)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Docker_Upd_Ipv6,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return fmt.Errorf("failed to update ipv6 option")
	}

	return nil
}

func (s *DockerMan) dockerOperation(hostID uint64, req model.DockerOperation) error {

	data, err := utils.ToJSONString(req)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Docker_Operation,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return fmt.Errorf("failed to execute operation")
	}

	return nil
}

func (s *DockerMan) inspect(hostID uint64, req model.Inspect) (*model.InspectResult, error) {
	result := model.InspectResult{
		Type: req.Type,
		ID:   req.ID,
	}

	data, err := utils.ToJSONString(req)
	if err != nil {
		return &result, err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Docker_Inspect,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return &result, err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return &result, fmt.Errorf("failed to inspect")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &result)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to inspect result: %v", err)
		return &result, fmt.Errorf("json err: %v", err)
	}

	return &result, nil
}

func (s *DockerMan) prune(hostID uint64, req model.Prune) (*model.PruneResult, error) {
	var result model.PruneResult

	data, err := utils.ToJSONString(req)
	if err != nil {
		return &result, err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Docker_Prune,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return &result, err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return &result, fmt.Errorf("failed to prune")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &result)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to prune result: %v", err)
		return &result, fmt.Errorf("json err: %v", err)
	}

	return &result, nil
}

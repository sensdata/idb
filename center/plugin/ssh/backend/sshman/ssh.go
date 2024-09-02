package sshman

import (
	"fmt"

	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/utils"
)

func (s *SSHMan) getSSHConfig(req model.SSHConfigReq) (*model.SSHInfo, error) {
	var sshInfo model.SSHInfo
	data, err := utils.ToJSONString(req)
	if err != nil {
		return &sshInfo, err
	}

	actionRequest := model.HostAction{
		HostID: req.HostID,
		Action: model.Action{
			Action: model.Ssh_Config,
			Data:   data,
		},
	}

	var actionResponse model.ActionResponse

	resp, err := s.restyClient.R().
		SetBody(actionRequest).
		SetResult(&actionResponse).
		Post("http://127.0.0.1:8080/idb/api/act/send")

	if err != nil {
		global.LOG.Error("failed to send request: %v", err)
		return &sshInfo, fmt.Errorf("failed to send request: %v", err)
	}

	if resp.StatusCode() != 200 {
		global.LOG.Error("failed to send request: %v", err)
		return &sshInfo, fmt.Errorf("received error response: %s", resp.Status())
	}

	global.LOG.Info("ssh result: %v", actionResponse)

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return &sshInfo, fmt.Errorf("failed to get ssh config")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &sshInfo)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to ssh config: %v", err)
		return &sshInfo, fmt.Errorf("json err: %v", err)
	}

	return &sshInfo, nil
}

func (s *SSHMan) updateSSH(req model.SSHUpdate) error {
	data, err := utils.ToJSONString(req)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: req.HostID,
		Action: model.Action{
			Action: model.Ssh_Config_Update,
			Data:   data,
		},
	}

	var actionResponse model.ActionResponse

	resp, err := s.restyClient.R().
		SetBody(actionRequest).
		SetResult(&actionResponse).
		Post("http://127.0.0.1:8080/idb/api/act/send")

	if err != nil {
		global.LOG.Error("failed to send request: %v", err)
		return fmt.Errorf("failed to send request: %v", err)
	}

	if resp.StatusCode() != 200 {
		global.LOG.Error("failed to send request: %v", err)
		return fmt.Errorf("received error response: %s", resp.Status())
	}

	global.LOG.Info("ssh result: %v", actionResponse)

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return fmt.Errorf("failed to update ssh config")
	}

	return nil
}

func (s *SSHMan) getSSHConfigContent(req model.SSHConfigReq) (*model.SSHConfigContent, error) {
	var sshContent model.SSHConfigContent
	data, err := utils.ToJSONString(req)
	if err != nil {
		return &sshContent, err
	}

	actionRequest := model.HostAction{
		HostID: req.HostID,
		Action: model.Action{
			Action: model.Ssh_Config_Content,
			Data:   data,
		},
	}

	var actionResponse model.ActionResponse

	resp, err := s.restyClient.R().
		SetBody(actionRequest).
		SetResult(&actionResponse).
		Post("http://127.0.0.1:8080/idb/api/act/send")

	if err != nil {
		global.LOG.Error("failed to send request: %v", err)
		return &sshContent, fmt.Errorf("failed to send request: %v", err)
	}

	if resp.StatusCode() != 200 {
		global.LOG.Error("failed to send request: %v", err)
		return &sshContent, fmt.Errorf("received error response: %s", resp.Status())
	}

	global.LOG.Info("ssh result: %v", actionResponse)

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return &sshContent, fmt.Errorf("failed to get ssh config content")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &sshContent)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to ssh config content: %v", err)
		return &sshContent, fmt.Errorf("json err: %v", err)
	}

	return &sshContent, nil
}

func (s *SSHMan) updateSSHContent(req model.ContentUpdate) error {
	data, err := utils.ToJSONString(req)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: req.HostID,
		Action: model.Action{
			Action: model.Ssh_Config_Content_Update,
			Data:   data,
		},
	}

	var actionResponse model.ActionResponse

	resp, err := s.restyClient.R().
		SetBody(actionRequest).
		SetResult(&actionResponse).
		Post("http://127.0.0.1:8080/idb/api/act/send")

	if err != nil {
		global.LOG.Error("failed to send request: %v", err)
		return fmt.Errorf("failed to send request: %v", err)
	}

	if resp.StatusCode() != 200 {
		global.LOG.Error("failed to send request: %v", err)
		return fmt.Errorf("received error response: %s", resp.Status())
	}

	global.LOG.Info("ssh result: %v", actionResponse)

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return fmt.Errorf("failed to update ssh config content")
	}

	return nil
}

func (s *SSHMan) operateSSH(req model.SSHOperate) error {
	data, err := utils.ToJSONString(req)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: req.HostID,
		Action: model.Action{
			Action: model.Ssh_Operate,
			Data:   data,
		},
	}

	var actionResponse model.ActionResponse

	resp, err := s.restyClient.R().
		SetBody(actionRequest).
		SetResult(&actionResponse).
		Post("http://127.0.0.1:8080/idb/api/act/send")

	if err != nil {
		global.LOG.Error("failed to send request: %v", err)
		return fmt.Errorf("failed to send request: %v", err)
	}

	if resp.StatusCode() != 200 {
		global.LOG.Error("failed to send request: %v", err)
		return fmt.Errorf("received error response: %s", resp.Status())
	}

	global.LOG.Info("ssh result: %v", actionResponse)

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return fmt.Errorf("failed to operate ssh service")
	}

	return nil
}

func (s *SSHMan) createKey(req model.GenerateKey) error {
	data, err := utils.ToJSONString(req)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: req.HostID,
		Action: model.Action{
			Action: model.Ssh_Secret_Create,
			Data:   data,
		},
	}

	var actionResponse model.ActionResponse

	resp, err := s.restyClient.R().
		SetBody(actionRequest).
		SetResult(&actionResponse).
		Post("http://127.0.0.1:8080/idb/api/act/send")

	if err != nil {
		global.LOG.Error("failed to send request: %v", err)
		return fmt.Errorf("failed to send request: %v", err)
	}

	if resp.StatusCode() != 200 {
		global.LOG.Error("failed to send request: %v", err)
		return fmt.Errorf("received error response: %s", resp.Status())
	}

	global.LOG.Info("ssh result: %v", actionResponse)

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return fmt.Errorf("failed to create ssh key")
	}

	return nil
}

func (s *SSHMan) listKeys(req model.ListKey) (*model.PageResult, error) {
	var pageResult model.PageResult
	data, err := utils.ToJSONString(req)
	if err != nil {
		return &pageResult, err
	}

	actionRequest := model.HostAction{
		HostID: req.HostID,
		Action: model.Action{
			Action: model.Ssh_Secret,
			Data:   data,
		},
	}

	var actionResponse model.ActionResponse

	resp, err := s.restyClient.R().
		SetBody(actionRequest).
		SetResult(&actionResponse).
		Post("http://127.0.0.1:8080/idb/api/act/send")

	if err != nil {
		global.LOG.Error("failed to send request: %v", err)
		return &pageResult, fmt.Errorf("failed to send request: %v", err)
	}

	if resp.StatusCode() != 200 {
		global.LOG.Error("failed to send request: %v", err)
		return &pageResult, fmt.Errorf("received error response: %s", resp.Status())
	}

	global.LOG.Info("ssh result: %v", actionResponse)

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return &pageResult, fmt.Errorf("failed to get ssh keys")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &pageResult)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to ssh keys: %v", err)
		return &pageResult, fmt.Errorf("json err: %v", err)
	}

	return &pageResult, nil
}

func (s *SSHMan) loadLog(req model.SearchSSHLog) (*model.SSHLog, error) {
	var sshLog model.SSHLog
	data, err := utils.ToJSONString(req)
	if err != nil {
		return &sshLog, err
	}

	actionRequest := model.HostAction{
		HostID: req.HostID,
		Action: model.Action{
			Action: model.Ssh_Log,
			Data:   data,
		},
	}

	var actionResponse model.ActionResponse

	resp, err := s.restyClient.R().
		SetBody(actionRequest).
		SetResult(&actionResponse).
		Post("http://127.0.0.1:8080/idb/api/act/send")

	if err != nil {
		global.LOG.Error("failed to send request: %v", err)
		return &sshLog, fmt.Errorf("failed to send request: %v", err)
	}

	if resp.StatusCode() != 200 {
		global.LOG.Error("failed to send request: %v", err)
		return &sshLog, fmt.Errorf("received error response: %s", resp.Status())
	}

	global.LOG.Info("ssh result: %v", actionResponse)

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return &sshLog, fmt.Errorf("failed to get ssh log")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &sshLog)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to ssh logs: %v", err)
		return &sshLog, fmt.Errorf("json err: %v", err)
	}

	return &sshLog, nil
}

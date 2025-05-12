package sshman

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/core/conn"
	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/utils"
)

func (s *SSHMan) sendAction(actionRequest model.HostAction) (*model.ActionResponse, error) {
	var actionResponse model.ActionResponse

	resp, err := s.restyClient.R().
		SetBody(actionRequest).
		SetResult(&actionResponse).
		Post("/actions") // 修改URL路径

	if err != nil {
		LOG.Error("failed to send request: %v", err)
		return nil, fmt.Errorf("failed to send request: %v", err)
	}

	if resp.StatusCode() != 200 {
		LOG.Error("received error response: %s", resp.Status())
		return nil, fmt.Errorf("received error response: %s", resp.Status())
	}

	return &actionResponse, nil
}

func (s *SSHMan) getSSHConfig(hostID uint64) (*model.SSHInfo, error) {
	var sshInfo model.SSHInfo

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Ssh_Config,
			Data:   "",
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return &sshInfo, err
	}

	if !actionResponse.Data.Action.Result {
		LOG.Error("action failed")
		return &sshInfo, fmt.Errorf("failed to get ssh config")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &sshInfo)
	if err != nil {
		LOG.Error("Error unmarshaling data to ssh config: %v", err)
		return &sshInfo, fmt.Errorf("json err: %v", err)
	}

	return &sshInfo, nil
}

func (s *SSHMan) updateSSH(hostID uint64, req model.SSHUpdate) error {
	data, err := utils.ToJSONString(req)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Ssh_Config_Update,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		LOG.Error("action failed")
		return fmt.Errorf("failed to update ssh config")
	}

	return nil
}

func (s *SSHMan) getSSHConfigContent(hostID uint64) (*model.SSHConfigContent, error) {
	var sshContent model.SSHConfigContent

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Ssh_Config_Content,
			Data:   "",
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return &sshContent, err
	}

	if !actionResponse.Data.Action.Result {
		LOG.Error("action failed")
		return &sshContent, fmt.Errorf("failed to get ssh config content")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &sshContent)
	if err != nil {
		LOG.Error("Error unmarshaling data to ssh config content: %v", err)
		return &sshContent, fmt.Errorf("json err: %v", err)
	}

	return &sshContent, nil
}

func (s *SSHMan) updateSSHContent(hostID uint64, req model.ContentUpdate) error {
	data, err := utils.ToJSONString(req)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Ssh_Config_Content_Update,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		LOG.Error("action failed")
		return fmt.Errorf("failed to update ssh config content")
	}

	return nil
}

func (s *SSHMan) operateSSH(hostID uint64, req model.SSHOperate) error {
	data, err := utils.ToJSONString(req)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Ssh_Operate,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		LOG.Error("action failed")
		return fmt.Errorf("failed to operate ssh service")
	}

	return nil
}

func (s *SSHMan) createKey(hostID uint64, req model.GenerateKey) error {
	data, err := utils.ToJSONString(req)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Ssh_Secret_Create,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		LOG.Error("action failed")
		return fmt.Errorf("failed to create ssh key")
	}

	return nil
}

func (s *SSHMan) listKeys(hostID uint64, req model.ListKey) (*model.PageResult, error) {
	var pageResult model.PageResult
	data, err := utils.ToJSONString(req)
	if err != nil {
		return &pageResult, err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Ssh_Secret,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return &pageResult, err
	}

	if !actionResponse.Data.Action.Result {
		LOG.Error("action failed")
		return &pageResult, fmt.Errorf("failed to get ssh keys")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &pageResult)
	if err != nil {
		LOG.Error("Error unmarshaling data to ssh keys: %v", err)
		return &pageResult, fmt.Errorf("json err: %v", err)
	}

	return &pageResult, nil
}

func (s *SSHMan) enableKey(hostID uint64, req model.EnableKey) error {
	data, err := utils.ToJSONString(req)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Ssh_Secret_Enable,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		LOG.Error("action failed")
		return fmt.Errorf("failed to enable ssh key")
	}

	return nil
}

func (s *SSHMan) removeKey(hostID uint64, req model.RemoveKey) error {
	data, err := utils.ToJSONString(req)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Ssh_Secret_Remove,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		LOG.Error("action failed")
		return fmt.Errorf("failed to remove ssh key")
	}

	return nil
}

func (s *SSHMan) downloadFile(c *gin.Context, hostID uint, path string) error {
	return conn.CENTER.DownloadFile(c, hostID, path)
}

func (s *SSHMan) setKeyPassword(hostID uint64, req model.SetKeyPassword) error {
	data, err := utils.ToJSONString(req)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Ssh_Set_Password,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		LOG.Error("action failed")
		return fmt.Errorf("failed to set key password")
	}

	return nil
}

func (s *SSHMan) updateKeyPassword(hostID uint64, req model.UpdateKeyPassword) error {
	data, err := utils.ToJSONString(req)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Ssh_Update_Password,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		LOG.Error("action failed")
		return fmt.Errorf("failed to update key password")
	}

	return nil
}

func (s *SSHMan) clearKeyPassword(hostID uint64, req model.SetKeyPassword) error {
	data, err := utils.ToJSONString(req)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Ssh_Clear_Password,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		LOG.Error("action failed")
		return fmt.Errorf("failed to clear key password")
	}

	return nil
}

func (s *SSHMan) loadLog(hostID uint64, req model.SearchSSHLog) (*model.SSHLog, error) {
	var sshLog model.SSHLog
	data, err := utils.ToJSONString(req)
	if err != nil {
		return &sshLog, err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Ssh_Log,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return &sshLog, err
	}

	if !actionResponse.Data.Action.Result {
		LOG.Error("action failed")
		return &sshLog, fmt.Errorf("failed to get ssh log")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &sshLog)
	if err != nil {
		LOG.Error("Error unmarshaling data to ssh logs: %v", err)
		return &sshLog, fmt.Errorf("json err: %v", err)
	}

	return &sshLog, nil
}

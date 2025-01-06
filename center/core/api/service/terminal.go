package service

import (
	"fmt"

	"github.com/sensdata/idb/center/core/conn"
	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/utils"
)

type TerminalService struct{}

type ITerminalService interface {
	Sessions(hostID uint) (*model.PageResult, error)
	Detach(hostID uint, req model.TerminalRequest) error
	Quit(hostID uint, req model.TerminalRequest) error
	Rename(hostID uint, req model.TerminalRequest) error
	Install(hostID uint) error
}

func NewITerminalService() ITerminalService {
	return &TerminalService{}
}

func (s *TerminalService) Sessions(hostID uint) (*model.PageResult, error) {
	var result model.PageResult

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Terminal_List,
			Data:   "",
		},
	}
	actionResponse, err := conn.CENTER.ExecuteAction(actionRequest)
	if err != nil {
		global.LOG.Error("Failed to send action %v", err)
		return &result, err
	}
	if !actionResponse.Result {
		global.LOG.Error("action failed")
		return &result, fmt.Errorf("failed to query sessions")
	}

	err = utils.FromJSONString(actionResponse.Data, &result)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to session list: %v", err)
		return &result, fmt.Errorf("json err: %v", err)
	}
	return &result, nil
}

func (s *TerminalService) Detach(hostID uint, req model.TerminalRequest) error {

	data, err := utils.ToJSONString(req)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Terminal_Detach,
			Data:   data,
		},
	}
	actionResponse, err := conn.CENTER.ExecuteAction(actionRequest)
	if err != nil {
		global.LOG.Error("Failed to send action %v", err)
		return err
	}
	if !actionResponse.Result {
		global.LOG.Error("action failed")
		return fmt.Errorf("failed to detach session")
	}

	return nil
}

func (s *TerminalService) Quit(hostID uint, req model.TerminalRequest) error {
	data, err := utils.ToJSONString(req)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Terminal_Finish,
			Data:   data,
		},
	}
	actionResponse, err := conn.CENTER.ExecuteAction(actionRequest)
	if err != nil {
		global.LOG.Error("Failed to send action %v", err)
		return err
	}
	if !actionResponse.Result {
		global.LOG.Error("action failed")
		return fmt.Errorf("failed to quit session")
	}

	return nil
}

func (s *TerminalService) Rename(hostID uint, req model.TerminalRequest) error {
	data, err := utils.ToJSONString(req)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Terminal_Rename,
			Data:   data,
		},
	}
	actionResponse, err := conn.CENTER.ExecuteAction(actionRequest)
	if err != nil {
		global.LOG.Error("Failed to send action %v", err)
		return err
	}
	if !actionResponse.Result {
		global.LOG.Error("action failed")
		return fmt.Errorf("failed to rename session")
	}

	return nil
}

func (s *TerminalService) Install(hostID uint) error {
	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Terminal_Install,
			Data:   "",
		},
	}
	actionResponse, err := conn.CENTER.ExecuteAction(actionRequest)
	if err != nil {
		global.LOG.Error("Failed to send action %v", err)
		return err
	}
	if !actionResponse.Result {
		global.LOG.Error("action failed")
		return fmt.Errorf("failed to install terminal")
	}

	return nil
}

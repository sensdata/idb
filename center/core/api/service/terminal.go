package service

import (
	"fmt"
	"time"

	"github.com/sensdata/idb/center/core/conn"
	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/message"
	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/utils"
)

type TerminalService struct{}

type ITerminalService interface {
	Sessions(hostID uint) (*model.PageResult, error)
	Prune(hostID uint) (*model.ScriptResult, error)
	Detach(token string, hostID uint, req model.TerminalRequest) error
	Quit(token string, hostID uint, req model.TerminalRequest) error
	Rename(token string, hostID uint, req model.TerminalRequest) error
	Install(hostID uint) (*model.ScriptResult, error)
}

func NewITerminalService() ITerminalService {
	return &TerminalService{}
}

func (s *TerminalService) Sessions(hostID uint) (*model.PageResult, error) {
	var result model.PageResult

	req := model.TerminalRequest{Type: string(message.SessionTypeScreen)}
	data, err := utils.ToJSONString(req)
	if err != nil {
		return &result, err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Terminal_List,
			Data:   data,
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

func (s *TerminalService) Prune(hostID uint) (*model.ScriptResult, error) {
	result := model.ScriptResult{
		LogPath: "",
		Start:   time.Now(),
		End:     time.Now(),
		Out:     "",
		Err:     "",
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Terminal_Prune,
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
		return &result, fmt.Errorf("failed to prune sessions")
	}

	err = utils.FromJSONString(actionResponse.Data, &result)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to script result: %v", err)
		return &result, fmt.Errorf("json err: %v", err)
	}

	return &result, nil
}

func (s *TerminalService) Detach(token string, hostID uint, req model.TerminalRequest) error {
	// 如果会话已经登记了token，说明正在被使用
	sessionToken, exist := conn.CENTER.GetSessionToken(req.Session)
	if exist && token != sessionToken {
		global.LOG.Error("session %s is being used by another user", req.Session)
		return fmt.Errorf("session %s is being used by another user", req.Session)
	}

	req.Type = string(message.SessionTypeScreen)
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
		return fmt.Errorf("operation failed")
	}
	if !actionResponse.Result {
		global.LOG.Error("action failed")
		global.LOG.Error("failed to detach session, might already been detached")
	}

	return nil
}

func (s *TerminalService) Quit(token string, hostID uint, req model.TerminalRequest) error {
	// 如果会话已经登记了token，说明正在被使用
	sessionToken, exist := conn.CENTER.GetSessionToken(req.Session)
	if exist && token != sessionToken {
		global.LOG.Error("session %s is being used by another user", req.Session)
		return fmt.Errorf("session %s is being used by another user", req.Session)
	}

	req.Type = string(message.SessionTypeScreen)
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
		return fmt.Errorf("operation failed")
	}
	if !actionResponse.Result {
		global.LOG.Error("failed to quit session, might already been quit")
	}

	return nil
}

func (s *TerminalService) Rename(token string, hostID uint, req model.TerminalRequest) error {
	// 如果会话已经登记了token，说明正在被使用
	sessionToken, exist := conn.CENTER.GetSessionToken(req.Session)
	if exist && token != sessionToken {
		global.LOG.Error("session %s is being used by another user", req.Session)
		return fmt.Errorf("session %s is being used by another user", req.Session)
	}

	req.Type = string(message.SessionTypeScreen)
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

func (s *TerminalService) Install(hostID uint) (*model.ScriptResult, error) {
	result := model.ScriptResult{
		LogPath: "",
		Start:   time.Now(),
		End:     time.Now(),
		Out:     "",
		Err:     "",
	}

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
		return &result, err
	}
	if !actionResponse.Result {
		global.LOG.Error("action failed")
		return &result, fmt.Errorf("failed to install terminal")
	}

	err = utils.FromJSONString(actionResponse.Data, &result)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to script result: %v", err)
		return &result, fmt.Errorf("json err: %v", err)
	}

	return &result, nil
}

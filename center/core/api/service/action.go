package service

import (
	"github.com/sensdata/idb/center/core/conn"
	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/model"
)

type ActionService struct{}

type IActionService interface {
	SendAction(action model.HostAction) (*model.HostAction, error)
}

func NewIActionService() IActionService {
	return &ActionService{}
}

func (s *ActionService) SendAction(action model.HostAction) (*model.HostAction, error) {
	result, err := conn.CENTER.ExecuteAction(action)
	if err != nil {
		global.LOG.Error("Failed to send action %v", err)
		return nil, err
	}

	global.LOG.Info("send action result: %v", result)
	return &model.HostAction{
		HostID: action.HostID,
		Action: *result,
	}, nil
}

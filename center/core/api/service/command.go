package service

import (
	"github.com/sensdata/idb/center/core/conn"
	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/model"
)

type CommandService struct{}

type ICommandService interface {
	SendCommand(command model.Command) (*model.CommandResult, error)
	SendCommandGroup(command model.CommandGroup) (*model.CommandGroupResult, error)
}

func NewICommandService() ICommandService {
	return &CommandService{}
}

func (s *CommandService) SendCommand(command model.Command) (*model.CommandResult, error) {
	result, err := conn.CENTER.ExecuteCommand(command)
	if err != nil {
		return nil, err
	}

	return &model.CommandResult{
		HostID: command.HostID,
		Result: result,
	}, nil
}

func (s *CommandService) SendCommandGroup(command model.CommandGroup) (*model.CommandGroupResult, error) {
	results, err := conn.CENTER.ExecuteCommandGroup(command)
	if err != nil {
		global.LOG.Error("Failed to send command group %v", err)
		return nil, err
	}

	global.LOG.Info("send command group result: %v", results)
	return &model.CommandGroupResult{
		HostID:  command.HostID,
		Results: results,
	}, nil
}

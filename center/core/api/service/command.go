package service

import (
	"github.com/sensdata/idb/center/core/api/dto"
	"github.com/sensdata/idb/center/core/conn"
	"github.com/sensdata/idb/center/global"
)

type CommandService struct{}

type ICommandService interface {
	SendCommand(command dto.Command) (*dto.CommandResult, error)
	SendCommandGroup(command dto.CommandGroup) (*dto.CommandGroupResult, error)
}

func NewICommandService() ICommandService {
	return &CommandService{}
}

func (s *CommandService) SendCommand(command dto.Command) (*dto.CommandResult, error) {
	result, err := conn.CENTER.ExecuteCommand(command)
	if err != nil {
		return nil, err
	}

	return &dto.CommandResult{
		HostID: command.HostID,
		Result: result,
	}, nil
}

func (s *CommandService) SendCommandGroup(command dto.CommandGroup) (*dto.CommandGroupResult, error) {
	results, err := conn.CENTER.ExecuteCommandGroup(command)
	if err != nil {
		global.LOG.Error("Failed to send command group %v", err)
		return nil, err
	}

	global.LOG.Info("send command group result: %v", results)
	return &dto.CommandGroupResult{
		HostID:  command.HostID,
		Results: results,
	}, nil
}

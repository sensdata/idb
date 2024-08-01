package service

import (
	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/core/api/dto"
	"github.com/sensdata/idb/center/core/conn"
)

type CommandService struct{}

type ICommandService interface {
	SendCommand(c *gin.Context, command dto.Command) (*dto.CommandResult, error)
	SendCommandGroup(c *gin.Context, command dto.CommandGroup) (*dto.CommandGroupResult, error)
}

func NewICommandService() ICommandService {
	return &CommandService{}
}

func (s *CommandService) SendCommand(c *gin.Context, command dto.Command) (*dto.CommandResult, error) {
	result, err := conn.CENTER.ExecuteCommand(command)
	if err != nil {
		return nil, err
	}

	return &dto.CommandResult{
		HostID: command.HostID,
		Result: result,
	}, nil
}

func (s *CommandService) SendCommandGroup(c *gin.Context, command dto.CommandGroup) (*dto.CommandGroupResult, error) {
	results, err := conn.CENTER.ExecuteCommandGroup(command)
	if err != nil {
		return nil, err
	}

	return &dto.CommandGroupResult{
		HostID:  command.HostID,
		Results: results,
	}, nil
}

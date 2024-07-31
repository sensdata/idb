package service

import (
	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/core/api/dto"
	"github.com/sensdata/idb/center/core/conn"
)

type CommandService struct{}

type ICommandService interface {
	SendCommand(c *gin.Context, command dto.Command) (*dto.CommandResult, error)
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

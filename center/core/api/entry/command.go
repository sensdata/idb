package entry

import (
	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/model"
)

func (b *BaseApi) SendCommand(c *gin.Context) {
	var req model.Command
	if err := CheckBindAndValidate(&req, c); err != nil {
		return
	}

	result, err := commandService.SendCommand(req)
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}
	SuccessWithData(c, result)
}

func (b *BaseApi) SendCommandGroup(c *gin.Context) {
	var req model.CommandGroup
	if err := CheckBindAndValidate(&req, c); err != nil {
		return
	}

	result, err := commandService.SendCommandGroup(req)
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}
	SuccessWithData(c, result)
}

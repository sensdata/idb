package entry

import (
	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/model"
)

// @Tags Command
// @Summary send linux command to host
// @Description 向目标设备发送linux指令
// @Accept json
// @Produce json
// @Param request body model.Command true "request"
// @Success 200
// @Router /commands [post]
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

// @Tags Command
// @Summary send group of linux commands to host
// @Description 向目标设备发送一组linux指令
// @Accept json
// @Produce json
// @Param request body model.CommandGroup true "request"
// @Success 200
// @Router /commands/group [post]
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

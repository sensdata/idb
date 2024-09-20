package entry

import (
	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/model"
)

// @Tags Action
// @Summary send action command to host
// @Description 像目标设备发送action指令
// @Accept json
// @Produce json
// @Param request body model.HostAction true "request"
// @Success 200 {object} model.HostAction
// @Router /act/send [post]
func (b *BaseApi) SendAction(c *gin.Context) {
	var req model.HostAction
	if err := CheckBindAndValidate(&req, c); err != nil {
		return
	}

	result, err := actionService.SendAction(req)
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}
	SuccessWithData(c, result)
}

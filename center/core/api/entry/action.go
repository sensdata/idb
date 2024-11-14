package entry

import (
	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/model"
)

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

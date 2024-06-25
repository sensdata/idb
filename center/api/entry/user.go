package entry

import (
	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/api/dto"
	"github.com/sensdata/idb/center/constant"
)

// List users
func (b *BaseApi) ListUser(c *gin.Context) {
	var req dto.ListUser
	if err := CheckBindAndValidate(&req, c); err != nil {
		return
	}

	result, err := userService.List(c, req)
	if err != nil {
		ErrorWithDetail(c, constant.CodeSuccess, constant.ErrNoRecords.Error(), err)
		return
	}
	SuccessWithData(c, result)
}

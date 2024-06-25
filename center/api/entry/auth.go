package entry

import (
	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/api/dto"
	"github.com/sensdata/idb/center/constant"
)

// Login handles user login and returns a JWT token
func (b *BaseApi) Login(c *gin.Context) {
	var req dto.Login
	if err := CheckBindAndValidate(&req, c); err != nil {
		return
	}

	result, err := authService.Login(c, req)

	if err != nil {
		ErrorWithDetail(c, constant.CodeErrUnauthorized, constant.ErrInternalServer.Error(), err)
		return
	}
	SuccessWithData(c, result)
}

func (b *BaseApi) Logout(c *gin.Context) {
	if err := authService.LogOut(c); err != nil {
		ErrorWithDetail(c, constant.CodeErrUnauthorized, constant.ErrInternalServer.Error(), err)
		return
	}
	SuccessWithData(c, nil)
}

package entry

import (
	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/model"
)

// @Tags Auth
// @Summary User login
// @Description 用户登录
// @Accept json
// @Produce json
// @Param request body model.Login true "request"
// @Success 200 {object} model.LoginResult
// @Router /auth/login [post]
func (b *BaseApi) Login(c *gin.Context) {
	var req model.Login
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

// @Tags Auth
// @Summary User logout
// @Description 用户登出
// @Accept json
// @Produce json
// @Param token header string true "Authentication token"
// @Success 200
// @Router /auth/logout [get]
func (b *BaseApi) Logout(c *gin.Context) {
	if err := authService.LogOut(c); err != nil {
		ErrorWithDetail(c, constant.CodeErrUnauthorized, constant.ErrInternalServer.Error(), err)
		return
	}
	SuccessWithData(c, nil)
}

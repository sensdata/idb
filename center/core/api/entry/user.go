package entry

import (
	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/model"
)

// @Tags User
// @Summary get user list
// @Description 获取用户列表
// @Accept json
// @Produce json
// @Param request body model.PageInfo true "request"
// @Success 200 {object} model.PageResult
// @Router /user/list [post]
func (b *BaseApi) ListUser(c *gin.Context) {
	var req model.PageInfo
	if err := CheckBindAndValidate(&req, c); err != nil {
		return
	}

	result, err := userService.List(req)
	if err != nil {
		ErrorWithDetail(c, constant.CodeSuccess, constant.ErrNoRecords.Error(), err)
		return
	}
	SuccessWithData(c, result)
}

// @Tags User
// @Summary create user
// @Description 新增用户
// @Accept json
// @Produce json
// @Param request body model.CreateUser true "request"
// @Success 200 {object} model.UserInfo
// @Router /user/create [post]
func (b *BaseApi) CreateUser(c *gin.Context) {
	var req model.CreateUser
	if err := CheckBindAndValidate(&req, c); err != nil {
		return
	}

	result, err := userService.Create(req)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	SuccessWithData(c, result)
}

// @Tags User
// @Summary update user
// @Description 更新用户
// @Accept json
// @Produce json
// @Param request body model.UpdateUser true "request"
// @Success 200
// @Router /user/update [post]
func (b *BaseApi) UpdateUser(c *gin.Context) {
	var req model.UpdateUser
	if err := CheckBindAndValidate(&req, c); err != nil {
		return
	}

	upMap := make(map[string]interface{})
	upMap["user_name"] = req.UserName
	upMap["group_id"] = req.GroupID
	upMap["valid"] = req.Valid
	if err := userService.Update(req.UserID, upMap); err != nil {
		ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	SuccessWithData(c, nil)
}

// @Tags User
// @Summary delete user
// @Description 删除用户
// @Accept json
// @Produce json
// @Param request body model.DeleteUser true "request"
// @Success 200
// @Router /user/delete [post]
func (b *BaseApi) DeleteUser(c *gin.Context) {
	var req model.DeleteUser
	if err := CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := userService.Delete([]uint{req.UserID}); err != nil {
		ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	SuccessWithData(c, nil)
}

// @Tags User
// @Summary valid user
// @Description 禁用/启用用户
// @Accept json
// @Produce json
// @Param request body model.ValidUser true "request"
// @Success 200
// @Router /user/valid [post]
func (b *BaseApi) ValidUser(c *gin.Context) {
	var req model.ValidUser
	if err := CheckBindAndValidate(&req, c); err != nil {
		return
	}

	upMap := make(map[string]interface{})
	upMap["valid"] = req.Valid
	if err := userService.Update(req.UserID, upMap); err != nil {
		ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	SuccessWithData(c, nil)
}

// @Tags User
// @Summary update password
// @Description 更新密码
// @Accept json
// @Produce json
// @Param request body model.ChangePassword true "request"
// @Success 200
// @Router /user/update/password [post]
func (b *BaseApi) ChangePassword(c *gin.Context) {
	var req model.ChangePassword
	if err := CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := userService.ChangePassword(req); err != nil {
		ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	SuccessWithData(c, nil)
}

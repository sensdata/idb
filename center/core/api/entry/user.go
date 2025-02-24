package entry

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/utils"
)

// @Tags User
// @Summary get user list
// @Description 获取用户列表
// @Accept json
// @Produce json
// @Param page query int true "Page number"
// @Param page_size query int true "Page size"
// @Success 200 {object} model.PageResult
// @Router /users [get]
func (b *BaseApi) ListUser(c *gin.Context) {
	var req model.PageInfo
	if err := CheckQueryAndValidate(&req, c); err != nil {
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
// @Router /users [post]
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
// @Router /users [put]
func (b *BaseApi) UpdateUser(c *gin.Context) {
	var req model.UpdateUser
	if err := CheckBindAndValidate(&req, c); err != nil {
		return
	}

	upMap := make(map[string]interface{})
	upMap["name"] = req.Name
	upMap["group_id"] = req.GroupID
	upMap["valid"] = req.Valid
	if err := userService.Update(req.ID, upMap); err != nil {
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
// @Param id query int true "User ID"
// @Success 200
// @Router /users [delete]
func (b *BaseApi) DeleteUser(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Query("id"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid user ID", err)
		return
	}

	if err := userService.Delete([]uint{uint(userID)}); err != nil {
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
// @Router /users/valid [put]
func (b *BaseApi) ValidUser(c *gin.Context) {
	var req model.ValidUser
	if err := CheckBindAndValidate(&req, c); err != nil {
		return
	}

	upMap := make(map[string]interface{})
	upMap["valid"] = req.Valid
	if err := userService.Update(req.ID, upMap); err != nil {
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
// @Router /users/password [put]
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

// @Tags User
// @Summary Get profile
// @Description Get profile
// @Accept json
// @Produce json
// @Success 200 {object} model.Profile
// @Router /users/profile [get]
func (b *BaseApi) Profile(c *gin.Context) {
	claimsInterface, exists := c.Get("user")
	if !exists {
		ErrorWithDetail(c, constant.CodeAuth, constant.ErrAuth.Error(), constant.ErrAuth)
		return
	}
	claims, ok := claimsInterface.(*utils.Claims)
	if !ok {
		ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), constant.ErrInternalServer)
		return
	}
	result, err := userService.Profile(claims.ID)
	if err != nil {
		ErrorWithDetail(c, constant.CodeSuccess, constant.ErrNoRecords.Error(), err)
		return
	}
	SuccessWithData(c, result)
}

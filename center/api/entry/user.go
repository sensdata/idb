package entry

import (
	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/api/dto"
	"github.com/sensdata/idb/center/constant"
)

// List users
func (b *BaseApi) ListUser(c *gin.Context) {
	var req dto.PageInfo
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

// Create user
func (b *BaseApi) CreateUser(c *gin.Context) {
	var req dto.CreateUser
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

// Update user
func (b *BaseApi) UpdateUser(c *gin.Context) {
	var req dto.UpdateUser
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

// Delete user
func (b *BaseApi) DeleteUser(c *gin.Context) {
	var req dto.DeleteUser
	if err := CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := userService.Delete([]uint{req.UserID}); err != nil {
		ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	SuccessWithData(c, nil)
}

// Valid user
func (b *BaseApi) ValidUser(c *gin.Context) {
	var req dto.ValidUser
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

// Change password
func (b *BaseApi) ChangePassword(c *gin.Context) {
	var req dto.ChangePassword
	if err := CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := userService.ChangePassword(req); err != nil {
		ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	SuccessWithData(c, nil)
}

package entry

import (
	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/api/dto"
	"github.com/sensdata/idb/center/constant"
)

// List groups
func (b *BaseApi) ListGroup(c *gin.Context) {
	var req dto.PageInfo
	if err := CheckBindAndValidate(&req, c); err != nil {
		return
	}

	result, err := groupService.List(req)
	if err != nil {
		ErrorWithDetail(c, constant.CodeSuccess, constant.ErrNoRecords.Error(), err)
		return
	}
	SuccessWithData(c, result)
}

// Create group
func (b *BaseApi) CreateGroup(c *gin.Context) {
	var req dto.CreateGroup
	if err := CheckBindAndValidate(&req, c); err != nil {
		return
	}

	result, err := groupService.Create(req)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	SuccessWithData(c, result)
}

// Update group
func (b *BaseApi) UpdateGroup(c *gin.Context) {
	var req dto.UpdateGroup
	if err := CheckBindAndValidate(&req, c); err != nil {
		return
	}

	upMap := make(map[string]interface{})
	upMap["group_name"] = req.GroupName
	if err := groupService.Update(req.GroupID, upMap); err != nil {
		ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	SuccessWithData(c, nil)
}

// Delete group
func (b *BaseApi) DeleteGroup(c *gin.Context) {
	var req dto.DeleteGroup
	if err := CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := groupService.Delete([]uint{req.GroupID}); err != nil {
		ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	SuccessWithData(c, nil)
}

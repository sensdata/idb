package entry

import (
	"github.com/gin-gonic/gin"

	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/model"
)

// @Tags Group
// @Summary get group list
// @Description 获取组列表
// @Accept json
// @Produce json
// @Param page query int true "Page number"
// @Param page_size query int true "Page size"
// @Success 200 {object} model.PageResult
// @Router /groups [get]
func (b *BaseApi) ListGroup(c *gin.Context) {
	var req model.PageInfo
	if err := CheckQueryAndValidate(&req, c); err != nil {
		return
	}

	result, err := groupService.List(req)
	if err != nil {
		ErrorWithDetail(c, constant.CodeSuccess, constant.ErrNoRecords.Error(), err)
		return
	}
	SuccessWithData(c, result)
}

// @Tags Group
// @Summary create group
// @Description 创建组
// @Accept json
// @Produce json
// @Param request body model.CreateGroup true "request"
// @Success 200 {object} model.GroupInfo
// @Router /groups [post]
func (b *BaseApi) CreateGroup(c *gin.Context) {
	var req model.CreateGroup
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

// @Tags Group
// @Summary update group
// @Description 更新组
// @Accept json
// @Produce json
// @Param id path int true "Group ID"
// @Param request body model.UpdateGroup true "request"
// @Success 200
// @Router /groups/{id} [put]
func (b *BaseApi) UpdateGroup(c *gin.Context) {
	var req model.UpdateGroup
	if err := CheckBindAndValidate(&req, c); err != nil {
		return
	}

	groupID, err := GetParamID(c)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid group ID", err)
		return
	}

	upMap := make(map[string]interface{})
	upMap["group_name"] = req.GroupName
	if err := groupService.Update(groupID, upMap); err != nil {
		ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	SuccessWithData(c, nil)
}

// @Tags Group
// @Summary delete group
// @Description 删除组
// @Accept json
// @Produce json
// @Param id path int true "Group ID"
// @Success 200
// @Router /groups/{id} [delete]
func (b *BaseApi) DeleteGroup(c *gin.Context) {
	groupID, err := GetParamID(c)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid group ID", err)
		return
	}

	if err := groupService.Delete([]uint{groupID}); err != nil {
		ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	SuccessWithData(c, nil)
}

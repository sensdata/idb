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
// @Param request body model.PageInfo true "request"
// @Success 200 {object} model.PageResult
// @Router /group/list [post]
func (b *BaseApi) ListGroup(c *gin.Context) {
	var req model.PageInfo
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

// @Tags Group
// @Summary create group
// @Description 创建组
// @Accept json
// @Produce json
// @Param request body model.CreateGroup true "request"
// @Success 200 {object} model.GroupInfo
// @Router /group/create [post]
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
// @Param request body model.UpdateGroup true "request"
// @Success 200
// @Router /group/update [post]
func (b *BaseApi) UpdateGroup(c *gin.Context) {
	var req model.UpdateGroup
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

// @Tags Group
// @Summary delete group
// @Description 删除组
// @Accept json
// @Produce json
// @Param request body model.UpdateGroup true "request"
// @Success 200
// @Router /group/update [post]
func (b *BaseApi) DeleteGroup(c *gin.Context) {
	var req model.DeleteGroup
	if err := CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := groupService.Delete([]uint{req.GroupID}); err != nil {
		ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	SuccessWithData(c, nil)
}

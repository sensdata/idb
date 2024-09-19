package entry

import (
	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/model"
)

// @Tags Host
// @Summary get host group list
// @Description 获取设备组列表
// @Accept json
// @Produce json
// @Param request body model.PageInfo true "request"
// @Success 200 {object} model.PageResult
// @Router /host/list/group [post]
func (b *BaseApi) ListHostGroup(c *gin.Context) {
	var req model.PageInfo
	if err := CheckBindAndValidate(&req, c); err != nil {
		return
	}

	result, err := hostService.ListGroup(req)
	if err != nil {
		ErrorWithDetail(c, constant.CodeSuccess, constant.ErrNoRecords.Error(), err)
		return
	}
	SuccessWithData(c, result)
}

// @Tags Host
// @Summary get host list
// @Description 获取设备组列表
// @Accept json
// @Produce json
// @Param request body model.ListHost true "request"
// @Success 200 {object} model.PageResult
// @Router /host/list [post]
func (b *BaseApi) ListHost(c *gin.Context) {
	var req model.ListHost
	if err := CheckBindAndValidate(&req, c); err != nil {
		return
	}

	result, err := hostService.List(req)
	if err != nil {
		ErrorWithDetail(c, constant.CodeSuccess, constant.ErrNoRecords.Error(), err)
		return
	}
	SuccessWithData(c, result)
}

// @Tags Host
// @Summary create host
// @Description 新增设备
// @Accept json
// @Produce json
// @Param request body model.CreateHost true "request"
// @Success 200 {object} model.HostInfo
// @Router /host/create [post]
func (b *BaseApi) CreateHost(c *gin.Context) {
	var req model.CreateHost
	if err := CheckBindAndValidate(&req, c); err != nil {
		return
	}

	result, err := hostService.Create(req)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	SuccessWithData(c, result)
}

// @Tags Host
// @Summary update host
// @Description 更新设备
// @Accept json
// @Produce json
// @Param request body model.UpdateHost true "request"
// @Success 200
// @Router /host/update [post]
func (b *BaseApi) UpdateHost(c *gin.Context) {
	var req model.UpdateHost
	if err := CheckBindAndValidate(&req, c); err != nil {
		return
	}

	upMap := make(map[string]interface{})
	upMap["name"] = req.Name
	upMap["group_id"] = req.GroupID
	if err := hostService.Update(req.HostID, upMap); err != nil {
		ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	SuccessWithData(c, nil)
}

// @Tags Host
// @Summary update host ssh config
// @Description 更新设备ssh配置
// @Accept json
// @Produce json
// @Param request body model.UpdateHostSSH true "request"
// @Success 200
// @Router /host/update/ssh [post]
func (b *BaseApi) UpdateHostSSH(c *gin.Context) {
	var req model.UpdateHostSSH
	if err := CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := hostService.UpdateSSH(req); err != nil {
		ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	SuccessWithData(c, nil)
}

// @Tags Host
// @Summary update host agent config
// @Description 更新设备agent配置
// @Accept json
// @Produce json
// @Param request body model.UpdateHostAgent true "request"
// @Success 200
// @Router /host/update/agent [post]
func (b *BaseApi) UpdateHostAgent(c *gin.Context) {
	var req model.UpdateHostAgent
	if err := CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := hostService.UpdateAgent(req); err != nil {
		ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	SuccessWithData(c, nil)
}

// @Tags Host
// @Summary test host ssh
// @Description 测试设备ssh
// @Accept json
// @Produce json
// @Param request body model.TestSSH true "request"
// @Success 200
// @Router /host/test/ssh [post]
func (b *BaseApi) TestHostSSH(c *gin.Context) {
	var req model.TestSSH
	if err := CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := hostService.TestSSH(req); err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}
	SuccessWithData(c, nil)
}

// @Tags Host
// @Summary test host agent
// @Description 测试设备agent
// @Accept json
// @Produce json
// @Param request body model.TestAgent true "request"
// @Success 200
// @Router /host/test/agent [post]
func (b *BaseApi) TestHostAgent(c *gin.Context) {
	var req model.TestAgent
	if err := CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := hostService.TestAgent(req); err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}
	SuccessWithData(c, nil)
}

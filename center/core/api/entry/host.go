package entry

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/model"
)

// @Tags Host
// @Summary get host group list
// @Description 获取设备组列表
// @Accept json
// @Produce json
// @Param page query int true "Page number"
// @Param page_size query int true "Page size"
// @Success 200 {object} model.PageResult
// @Router /hosts/groups [get]
func (b *BaseApi) ListHostGroup(c *gin.Context) {
	var req model.PageInfo
	if err := CheckQueryAndValidate(&req, c); err != nil {
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
// @Param page query int true "Page number"
// @Param page_size query int true "Page size"
// @Param group_id query int false "Group ID"
// @Param keyword query string false "Keyword"
// @Success 200 {object} model.PageResult
// @Router /hosts [get]
func (b *BaseApi) ListHost(c *gin.Context) {
	var req model.ListHost
	if err := CheckQueryAndValidate(&req, c); err != nil {
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
// @Router /hosts [post]
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
// @Param id path int true "Host ID"
// @Param request body model.UpdateHost true "request"
// @Success 200
// @Router /hosts [put]
func (b *BaseApi) UpdateHost(c *gin.Context) {
	var req model.UpdateHost
	if err := CheckBindAndValidate(&req, c); err != nil {
		return
	}

	upMap := make(map[string]interface{})
	upMap["name"] = req.Name
	upMap["group_id"] = req.GroupID
	if err := hostService.Update(req.ID, upMap); err != nil {
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
// @Param host path int true "Host ID"
// @Param request body model.UpdateHostSSH true "request"
// @Success 200
// @Router /hosts/{host}/ssh [put]
func (b *BaseApi) UpdateHostSSH(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	var req model.UpdateHostSSH
	if err := CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := hostService.UpdateSSH(uint(hostID), req); err != nil {
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
// @Param host path int true "Host ID"
// @Param request body model.UpdateHostAgent true "request"
// @Success 200
// @Router /hosts/{host}/agent [put]
func (b *BaseApi) UpdateHostAgent(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	var req model.UpdateHostAgent
	if err := CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := hostService.UpdateAgent(uint(hostID), req); err != nil {
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
// @Param host path int true "Host ID"
// @Param request body model.TestSSH true "request"
// @Success 200
// @Router /hosts/{host}/test/ssh [post]
func (b *BaseApi) TestHostSSH(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	var req model.TestSSH
	if err := CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := hostService.TestSSH(uint(hostID), req); err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}
	SuccessWithData(c, nil)
}

// @Tags Host
// @Summary test host agent
// @Description 测试设备 agent
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param request body model.TestAgent true "request"
// @Success 200
// @Router /hosts/{host}/test/agent [post]
func (b *BaseApi) TestHostAgent(c *gin.Context) {

	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	var req model.TestAgent
	if err := CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := hostService.TestAgent(uint(hostID), req); err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}
	SuccessWithData(c, nil)
}

// @Tags Host
// @Summary install agent on host
// @Description 在设备上安装 agent
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Success 200
// @Router /hosts/{host}/install/agent [post]
func (b *BaseApi) InstallAgent(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	if err := hostService.InstallAgent(uint(hostID)); err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}
	SuccessWithData(c, nil)
}

package entry

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/model"
)

// @Tags Host
// @Summary get host group list
// @Description get host group list
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
// @Description get host list
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
// @Summary Add host
// @Description Add host
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
// @Summary Update host
// @Description Update host
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
// @Summary Delete host
// @Description Delete host
// @Accept json
// @Produce json
// @Param id query int true "Host ID"
// @Success 200
// @Router /hosts [delete]
func (b *BaseApi) DeleteHost(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Query("id"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	if err := hostService.Delete(uint(hostID)); err != nil {
		ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	SuccessWithData(c, nil)
}

// @Tags Host
// @Summary Get host info
// @Description Get host info
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Success 200 {object} model.HostInfo
// @Router /hosts/{host} [get]
func (b *BaseApi) HostInfo(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	info, err := hostService.Info(uint(hostID))
	if err != nil {
		ErrorWithDetail(c, constant.CodeSuccess, constant.ErrNoRecords.Error(), err)
		return
	}
	SuccessWithData(c, info)
}

// @Tags Host
// @Summary Get host status
// @Description Get host status
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Success 200 {object} model.HostStatus
// @Router /hosts/{host}/status [get]
func (b *BaseApi) HostStatus(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	status, err := hostService.Status(uint(hostID))
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}
	SuccessWithData(c, status)
}

// @Tags Host
// @Summary Update ssh config in host
// @Description Update ssh config in host
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param request body model.UpdateHostSSH true "request"
// @Success 200
// @Router /hosts/{host}/conf/ssh [put]
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
// @Summary Update agent config of host
// @Description Update agent config of host
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param request body model.UpdateHostAgent true "request"
// @Success 200
// @Router /hosts/{host}/conf/agent [put]
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
// @Summary Test ssh connection to host
// @Description Test ssh connection to host
// @Accept json
// @Produce json
// @Param request body model.TestSSH true "request"
// @Success 200
// @Router /hosts/test/ssh [post]
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
// @Summary Test agent connection to host
// @Description Test agent connection to host
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
// @Summary Get agent status in host
// @Description Get agent status in host
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Success 200 {object} model.AgentStatus
// @Router /hosts/{host}/agent/status [get]
func (b *BaseApi) AgentStatus(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	status, err := hostService.AgentStatus(uint(hostID))
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}
	SuccessWithData(c, status)
}

// @Tags Host
// @Summary Install agent in host
// @Description Install agent in host
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Success 200
// @Router /hosts/{host}/agent/install [post]
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

// @Tags Host
// @Summary Restart agent
// @Description Restart agent
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Success 200
// @Router /hosts/{host}/agent/restart [post]
func (b *BaseApi) RestartAgent(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	err = hostService.RestartAgent(uint(hostID))
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}
	SuccessWithData(c, nil)
}

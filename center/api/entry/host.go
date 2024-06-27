package entry

import (
	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/api/dto"
	"github.com/sensdata/idb/center/constant"
)

// List host group
func (b *BaseApi) ListHostGroup(c *gin.Context) {
	var req dto.PageInfo
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

// List hosts
func (b *BaseApi) ListHost(c *gin.Context) {
	var req dto.ListHost
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

// Create host
func (b *BaseApi) CreateHost(c *gin.Context) {
	var req dto.CreateHost
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

// Update host
func (b *BaseApi) UpdateHost(c *gin.Context) {
	var req dto.UpdateHost
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

// Update host ssh
func (b *BaseApi) UpdateHostSSH(c *gin.Context) {
	var req dto.UpdateHostSSH
	if err := CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := hostService.UpdateSSH(req); err != nil {
		ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	SuccessWithData(c, nil)
}

// Update host agent
func (b *BaseApi) UpdateHostAgent(c *gin.Context) {
	var req dto.UpdateHostAgent
	if err := CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := hostService.UpdateAgent(req); err != nil {
		ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	SuccessWithData(c, nil)
}

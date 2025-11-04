package entry

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/core/plugin"
	"github.com/sensdata/idb/center/core/plugin/shared"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/model"
)

func getPma() (shared.Pma, error) {
	// 获取插件客户端
	plugin, err := plugin.PLUGINSERVER.GetPlugin("pma")
	if err != nil {
		return nil, err
	}
	// 类型断言为 gRPC client
	client, ok := plugin.Stub.(shared.Pma)
	if !ok {
		return nil, errors.New("invalid plugin client")
	}
	return client, nil
}

// @Tags Pma
// @Summary List pma compose
// @Description List pma compose
// @Accept json
// @Produce json
// @Param host path string true "host"
// @Param page query uint true "Page"
// @Param page_size query uint true "Page size"
// @Success 200 {object} model.GetComposesResponse
// @Router /pma/{host} [get]
func (a *BaseApi) PmaComposes(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	page, err := strconv.ParseInt(c.Query("page"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid page", err)
		return
	}

	pageSize, err := strconv.ParseInt(c.Query("page_size"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid page_size", err)
		return
	}

	req := model.GetComposesRequest{
		Page:     int(page),
		PageSize: int(pageSize),
	}

	// 获取插件
	client, err := getPma()
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}
	// 调用插件方法
	resp, err := client.GetComposes(hostID, req)
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	SuccessWithData(c, resp)
}

// @Tags Pma
// @Summary Operation pma compose
// @Description Operation pma compose
// @Accept json
// @Produce json
// @Param host path string true "host"
// @Param req body model.OperateRequest true "req"
// @Success 200
// @Router /pma/{host}/operation [post]
func (a *BaseApi) PmaOperation(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	var req model.OperateRequest
	if err := CheckBindAndValidate(&req, c); err != nil {
		return
	}

	// 获取插件
	client, err := getPma()
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	// 调用插件方法
	err = client.Operation(hostID, req)
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	SuccessWithData(c, nil)
}

// @Tags Pma
// @Summary Set pma port
// @Description Set pma port
// @Accept json
// @Produce json
// @Param host path string true "host"
// @Param req body model.SetPortRequest true "req"
// @Success 200
// @Router /pma/{host}/port [post]
func (a *BaseApi) PmaSetPort(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	var req model.SetPortRequest
	if err := CheckBindAndValidate(&req, c); err != nil {
		return
	}

	// 获取插件
	client, err := getPma()
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	// 调用插件方法
	err = client.SetPort(hostID, req)
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	SuccessWithData(c, nil)
}

// @Tags Pma
// @Summary List pma servers
// @Description List pma servers
// @Accept json
// @Produce json
// @Param host path string true "host"
// @Param name query string true "name"
// @Param page query uint true "Page"
// @Param page_size query uint true "Page size"
// @Success 200 {object} model.GetServersResponse
// @Router /pma/{host}/servers [get]
func (a *BaseApi) PmaGetServers(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	var req model.GetServersRequest
	if err := CheckQueryAndValidate(&req, c); err != nil {
		return
	}

	// 获取插件
	client, err := getPma()
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	// 调用插件方法
	resp, err := client.GetServers(hostID, req)
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	SuccessWithData(c, resp)
}

// @Tags Pma
// @Summary Add pma server
// @Description Add pma server
// @Accept json
// @Produce json
// @Param host path string true "host"
// @Param req body model.AddServerRequest true "req"
// @Success 200
// @Router /pma/{host}/server [post]
func (a *BaseApi) PmaAddServer(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	var req model.AddServerRequest
	if err := CheckBindAndValidate(&req, c); err != nil {
		return
	}

	// 获取插件
	client, err := getPma()
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	// 调用插件方法
	err = client.AddServer(hostID, req)
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	SuccessWithData(c, nil)
}

// @Tags Pma
// @Summary Update pma server
// @Description Update pma server
// @Accept json
// @Produce json
// @Param host path string true "host"
// @Param req body model.AddServerRequest true "req"
// @Success 200
// @Router /pma/{host}/server [put]
func (a *BaseApi) PmaUpdateServer(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	var req model.AddServerRequest
	if err := CheckBindAndValidate(&req, c); err != nil {
		return
	}

	// 获取插件
	client, err := getPma()
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	// 调用插件方法
	err = client.UpdateServer(hostID, req)
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	SuccessWithData(c, nil)
}

// @Tags Pma
// @Summary Delete pma server
// @Description Delete pma server
// @Accept json
// @Produce json
// @Param host path string true "host"
// @Param name query string true "name"
// @Param host query string true "host"
// @Param port query string true "port"
// @Success 200
// @Router /pma/{host}/server [delete]
func (a *BaseApi) PmaRemoveServer(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	var req model.RemoveServerRequest
	if err := CheckQueryAndValidate(&req, c); err != nil {
		return
	}

	// 获取插件
	client, err := getPma()
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	// 调用插件方法
	err = client.RemoveServer(hostID, req)
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	SuccessWithData(c, nil)
}

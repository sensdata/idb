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

func getMysqlManager() (shared.MysqlManager, error) {
	// 获取插件客户端
	plugin, err := plugin.PLUGINSERVER.GetPlugin("mysqlmanager")
	if err != nil {
		return nil, err
	}
	// 类型断言为 gRPC client
	client, ok := plugin.Stub.(shared.MysqlManager)
	if !ok {
		return nil, errors.New("invalid plugin client")
	}
	return client, nil
}

// @Tags Mysql
// @Summary List mysql compose
// @Description List mysql compose
// @Accept json
// @Produce json
// @Param host path string true "host"
// @Param page query uint true "Page"
// @Param page_size query uint true "Page size"
// @Success 200 {object} model.GetComposesResponse
// @Router /mysql/{host} [get]
func (a *BaseApi) MysqlComposes(c *gin.Context) {
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
	client, err := getMysqlManager()
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

// @Tags Mysql
// @Summary Operation mysql compose
// @Description Operation mysql compose
// @Accept json
// @Produce json
// @Param host path string true "host"
// @Param req body model.OperateRequest true "req"
// @Success 200
// @Router /mysql/{host}/operation [post]
func (a *BaseApi) MysqlOperation(c *gin.Context) {
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
	client, err := getMysqlManager()
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

// @Tags Mysql
// @Summary Set mysql port
// @Description Set mysql port
// @Accept json
// @Produce json
// @Param host path string true "host"
// @Param req body model.SetPortRequest true "req"
// @Success 200
// @Router /mysql/{host}/port [post]
func (a *BaseApi) MysqlSetPort(c *gin.Context) {
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
	client, err := getMysqlManager()
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

// @Tags Mysql
// @Summary Get mysql conf
// @Description Get mysql conf
// @Accept json
// @Produce json
// @Param host path string true "host"
// @Param name query string true "name"
// @Success 200 {object} model.GetConfResponse
// @Router /mysql/{host}/conf [get]
func (a *BaseApi) MysqlGetConf(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	name := c.Query("name")
	if name == "" {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid name", nil)
		return
	}

	req := model.GetConfRequest{
		Name: name,
	}

	// 获取插件
	client, err := getMysqlManager()
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	// 调用插件方法
	resp, err := client.GetConf(hostID, req)
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	SuccessWithData(c, resp)
}

// @Tags Mysql
// @Summary Set mysql conf
// @Description Set mysql conf
// @Accept json
// @Produce json
// @Param host path string true "host"
// @Param req body model.SetConfRequest true "req"
// @Success 200
// @Router /mysql/{host}/conf [post]
func (a *BaseApi) MysqlSetConf(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	var req model.SetConfRequest
	if err := CheckBindAndValidate(&req, c); err != nil {
		return
	}

	// 获取插件
	client, err := getMysqlManager()
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	// 调用插件方法
	err = client.SetConf(hostID, req)
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	SuccessWithData(c, nil)
}

// @Tags Mysql
// @Summary Get mysql remote access
// @Description Get mysql remote access
// @Accept json
// @Produce json
// @Param host path string true "host"
// @Param name query string true "name"
// @Success 200 {object} model.GetRemoteAccessResponse
// @Router /mysql/{host}/remote_access [get]
func (a *BaseApi) MysqlGetRemoteAccess(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	name := c.Query("name")
	if name == "" {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid name", nil)
		return
	}

	req := model.GetRemoteAccessRequest{
		Name: name,
	}

	// 获取插件
	client, err := getMysqlManager()
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	// 调用插件方法
	resp, err := client.GetRemoteAccess(hostID, req)
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	SuccessWithData(c, resp)
}

// @Tags Mysql
// @Summary Set mysql remote access
// @Description Set mysql remote access
// @Accept json
// @Produce json
// @Param host path string true "host"
// @Param req body model.SetRemoteAccessRequest true "req"
// @Success 200
// @Router /mysql/{host}/remote_access [post]
func (a *BaseApi) MysqlSetRemoteAccess(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	var req model.SetRemoteAccessRequest
	if err := CheckBindAndValidate(&req, c); err != nil {
		return
	}

	// 获取插件
	client, err := getMysqlManager()
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	// 调用插件方法
	err = client.SetRemoteAccess(hostID, req)
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	SuccessWithData(c, nil)
}

// @Tags Mysql
// @Summary Get mysql root password
// @Description Get mysql root password
// @Accept json
// @Produce json
// @Param host path string true "host"
// @Param name query string true "name"
// @Success 200 {object} model.GetRootPasswordResponse
// @Router /mysql/{host}/password [get]
func (a *BaseApi) MysqlGetRootPassword(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	name := c.Query("name")
	if name == "" {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid name", nil)
		return
	}

	req := model.GetRootPasswordRequest{
		Name: name,
	}

	// 获取插件
	client, err := getMysqlManager()
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	// 调用插件方法
	resp, err := client.GetRootPassword(hostID, req)
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	SuccessWithData(c, resp)
}

// @Tags Mysql
// @Summary Set mysql root password
// @Description Set mysql root password
// @Accept json
// @Produce json
// @Param host path string true "host"
// @Param req body model.SetRootPasswordRequest true "req"
// @Success 200
// @Router /mysql/{host}/password [post]
func (a *BaseApi) MysqlSetRootPassword(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	var req model.SetRootPasswordRequest
	if err := CheckBindAndValidate(&req, c); err != nil {
		return
	}

	// 获取插件
	client, err := getMysqlManager()
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	// 调用插件方法
	err = client.SetRootPassword(hostID, req)
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	SuccessWithData(c, nil)
}

// @Tags Mysql
// @Summary Get mysql connection info
// @Description Get mysql connection info
// @Accept json
// @Produce json
// @Param host path string true "host"
// @Param name query string true "name"
// @Success 200 {object} model.GetConnectionInfoResponse
// @Router /mysql/{host}/connection [get]
func (a *BaseApi) MysqlGetConnectionInfo(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	name := c.Query("name")
	if name == "" {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid name", nil)
		return
	}

	req := model.GetConnectionInfoRequest{
		Name: name,
	}

	// 获取插件
	client, err := getMysqlManager()
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	// 调用插件方法
	resp, err := client.GetConnectionInfo(hostID, req)
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	SuccessWithData(c, resp)
}

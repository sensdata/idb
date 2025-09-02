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

func getRedis() (shared.Redis, error) {
	// 获取插件客户端
	plugin, err := plugin.PLUGINSERVER.GetPlugin("redis")
	if err != nil {
		return nil, err
	}
	// 类型断言为 gRPC client
	client, ok := plugin.Stub.(shared.Redis)
	if !ok {
		return nil, errors.New("invalid plugin client")
	}
	return client, nil
}

// @Tags Redis
// @Summary List redis compose
// @Description List redis compose
// @Accept json
// @Produce json
// @Param host path string true "host"
// @Param page query uint true "Page"
// @Param page_size query uint true "Page size"
// @Success 200 {object} model.GetComposesResponse
// @Router /redis/{host} [get]
func (a *BaseApi) RedisComposes(c *gin.Context) {
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
	client, err := getRedis()
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

// @Tags Redis
// @Summary Operation redis compose
// @Description Operation redis compose
// @Accept json
// @Produce json
// @Param host path string true "host"
// @Param req body model.OperateRequest true "req"
// @Success 200
// @Router /redis/{host}/operation [post]
func (a *BaseApi) RedisOperation(c *gin.Context) {
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
	client, err := getRedis()
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

// @Tags Redis
// @Summary Set redis port
// @Description Set redis port
// @Accept json
// @Produce json
// @Param host path string true "host"
// @Param req body model.SetPortRequest true "req"
// @Success 200
// @Router /redis/{host}/port [post]
func (a *BaseApi) RedisSetPort(c *gin.Context) {
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
	client, err := getRedis()
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

// @Tags Redis
// @Summary Get redis conf
// @Description Get redis conf
// @Accept json
// @Produce json
// @Param host path string true "host"
// @Param name query string true "name"
// @Success 200 {object} model.GetConfResponse
// @Router /redis/{host}/conf [get]
func (a *BaseApi) RedisGetConf(c *gin.Context) {
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
	client, err := getRedis()
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

// @Tags Redis
// @Summary Set redis conf
// @Description Set redis conf
// @Accept json
// @Produce json
// @Param host path string true "host"
// @Param req body model.SetConfRequest true "req"
// @Success 200
// @Router /redis/{host}/conf [post]
func (a *BaseApi) RedisSetConf(c *gin.Context) {
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
	client, err := getRedis()
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

// @Tags Redis
// @Summary Get redis remote access
// @Description Get redis remote access
// @Accept json
// @Produce json
// @Param host path string true "host"
// @Param name query string true "name"
// @Success 200 {object} model.GetRemoteAccessResponse
// @Router /redis/{host}/remote_access [get]
func (a *BaseApi) RedisGetRemoteAccess(c *gin.Context) {
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
	client, err := getRedis()
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

// @Tags Redis
// @Summary Set redis remote access
// @Description Set redis remote access
// @Accept json
// @Produce json
// @Param host path string true "host"
// @Param req body model.SetRemoteAccessRequest true "req"
// @Success 200
// @Router /redis/{host}/remote_access [post]
func (a *BaseApi) RedisSetRemoteAccess(c *gin.Context) {
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
	client, err := getRedis()
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

// @Tags Redis
// @Summary Get redis root password
// @Description Get redis root password
// @Accept json
// @Produce json
// @Param host path string true "host"
// @Param name query string true "name"
// @Success 200 {object} model.GetRootPasswordResponse
// @Router /redis/{host}/password [get]
func (a *BaseApi) RedisGetRootPassword(c *gin.Context) {
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
	client, err := getRedis()
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

// @Tags Redis
// @Summary Set redis root password
// @Description Set redis root password
// @Accept json
// @Produce json
// @Param host path string true "host"
// @Param req body model.SetRootPasswordRequest true "req"
// @Success 200
// @Router /redis/{host}/password [post]
func (a *BaseApi) RedisSetRootPassword(c *gin.Context) {
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
	client, err := getRedis()
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

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

func getPostgreSql() (shared.PostgreSql, error) {
	// 获取插件客户端
	plugin, err := plugin.PLUGINSERVER.GetPlugin("postgresql")
	if err != nil {
		return nil, err
	}
	// 类型断言为 gRPC client
	client, ok := plugin.Stub.(shared.PostgreSql)
	if !ok {
		return nil, errors.New("invalid plugin client")
	}
	return client, nil
}

// @Tags PostgreSql
// @Summary List postgresql compose
// @Description List postgresql compose
// @Accept json
// @Produce json
// @Param host path string true "host"
// @Param page query uint true "Page"
// @Param page_size query uint true "Page size"
// @Success 200 {object} model.GetComposesResponse
// @Router /postgresql/{host} [get]
func (a *BaseApi) PostgreSqlComposes(c *gin.Context) {
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
	client, err := getPostgreSql()
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

// @Tags PostgreSql
// @Summary Operation postgresql compose
// @Description Operation postgresql compose
// @Accept json
// @Produce json
// @Param host path string true "host"
// @Param req body model.OperateRequest true "req"
// @Success 200
// @Router /postgresql/{host}/operation [post]
func (a *BaseApi) PostgreSqlOperation(c *gin.Context) {
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
	client, err := getPostgreSql()
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

// @Tags PostgreSql
// @Summary Set postgresql port
// @Description Set postgresql port
// @Accept json
// @Produce json
// @Param host path string true "host"
// @Param req body model.SetPortRequest true "req"
// @Success 200
// @Router /postgresql/{host}/port [post]
func (a *BaseApi) PostgreSqlSetPort(c *gin.Context) {
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
	client, err := getPostgreSql()
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

// @Tags PostgreSql
// @Summary Get postgresql conf
// @Description Get postgresql conf
// @Accept json
// @Produce json
// @Param host path string true "host"
// @Param name query string true "name"
// @Success 200 {object} model.GetConfResponse
// @Router /postgresql/{host}/conf [get]
func (a *BaseApi) PostgreSqlGetConf(c *gin.Context) {
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
	client, err := getPostgreSql()
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

// @Tags PostgreSql
// @Summary Set postgresql conf
// @Description Set postgresql conf
// @Accept json
// @Produce json
// @Param host path string true "host"
// @Param req body model.SetConfRequest true "req"
// @Success 200
// @Router /postgresql/{host}/conf [post]
func (a *BaseApi) PostgreSqlSetConf(c *gin.Context) {
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
	client, err := getPostgreSql()
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

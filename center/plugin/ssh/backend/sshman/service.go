package sshman

import (
	_ "embed"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/sensdata/idb/center/core/api"
	"github.com/sensdata/idb/center/core/conn"
	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/helper"
	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/plugin"
	"gopkg.in/yaml.v2"
)

type SSHMan struct {
	config      plugin.PluginConfig
	restyClient *resty.Client
}

var Plugin = SSHMan{}

//go:embed plug.yaml
var plugYAML []byte

func (s *SSHMan) Initialize() {
	global.LOG.Info("sshman init begin")

	if err := yaml.Unmarshal(plugYAML, &s.config); err != nil {
		global.LOG.Error("Failed to load fileman yaml: %v", err)
		return
	}

	baseUrl := fmt.Sprintf("http://%s:%d/api/v1", "127.0.0.1", conn.CONFMAN.GetConfig().Port)
	s.restyClient = resty.New().
		SetBaseURL(baseUrl).
		SetHeader("Content-Type", "application/json")

	api.API.SetUpPluginRouters(
		"ssh",
		[]plugin.PluginRoute{
			{Method: "GET", Path: "/info", Handler: s.GetPluginInfo},
			{Method: "GET", Path: "/menu", Handler: s.GetMenu},
			{Method: "GET", Path: "/config/:host_id", Handler: s.GetSSHConfig},
			{Method: "PUT", Path: "/config/:host_id", Handler: s.UpdateSSHConfig},
			{Method: "GET", Path: "/config/content/:host_id", Handler: s.GetSSHConfigContent},
			{Method: "PUT", Path: "/config/content/:host_id", Handler: s.UpdateSSHConfigContent},
			{Method: "POST", Path: "/operate/:host_id", Handler: s.OperateSSH},
			{Method: "POST", Path: "/keys/:host_id", Handler: s.CreateKey},
			{Method: "GET", Path: "/keys/:host_id", Handler: s.ListKey},
			{Method: "GET", Path: "/logs/:host_id", Handler: s.LoadSSHLogs},
		},
	)
	global.LOG.Info("sshman init end")
}

func (s *SSHMan) Release() {

}

// @Tags SSH
// @Summary Get plugin info
// @Description Get plugin information
// @Accept json
// @Produce json
// @Success 200 {object} plugin.PluginInfo
// @Router /ssh/info [get]
func (s *SSHMan) GetPluginInfo(c *gin.Context) {
	pluginInfo, err := s.getPluginInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"type": "info", "payload": pluginInfo})
}

// @Tags SSH
// @Summary Get plugin menu
// @Description Get plugin menu items
// @Accept json
// @Produce json
// @Success 200 {array} plugin.MenuItem
// @Router /ssh/menu [get]
func (s *SSHMan) GetMenu(c *gin.Context) {
	menuItems, err := s.getMenus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"type": "menu", "payload": menuItems})
}

func (s *SSHMan) getPluginInfo() (plugin.PluginInfo, error) {
	return s.config.Plugin, nil
}

func (s *SSHMan) getMenus() ([]plugin.MenuItem, error) {
	return s.config.Menu, nil
}

// @Tags SSH
// @Summary Get SSH configurations on host
// @Description Get SSH configurations on host
// @Accept json
// @Produce json
// @Param host_id path uint true "Host ID"
// @Success 200 {object} model.SSHInfo
// @Router /ssh/config/{host_id} [get]
func (s *SSHMan) GetSSHConfig(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host_id"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host_id", err)
		return
	}

	req := model.SSHConfigReq{
		HostID: uint(hostID),
	}

	info, err := s.getSSHConfig(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithData(c, info)
}

// @Tags SSH
// @Summary Update SSH configurations on host
// @Description Update SSH configurations on host
// @Accept json
// @Produce json
// @Param host_id path uint true "Host ID"
// @Param request body model.SSHUpdate true "request"
// @Success 200
// @Router /ssh/config/{host_id} [put]
func (s *SSHMan) UpdateSSHConfig(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host_id"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host_id", err)
		return
	}

	var req model.SSHUpdate
	if err := helper.CheckBind(&req, c); err != nil {
		return
	}
	req.HostID = uint(hostID)

	if err := s.updateSSH(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags SSH
// @Summary Get SSH config file content on host
// @Description Get SSH config file content on host
// @Accept json
// @Produce json
// @Param host_id path uint true "Host ID"
// @Success 200 {object} model.SSHConfigContent
// @Router /ssh/config/content/{host_id} [get]
func (s *SSHMan) GetSSHConfigContent(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host_id"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host_id", err)
		return
	}

	req := model.SSHConfigReq{
		HostID: uint(hostID),
	}

	info, err := s.getSSHConfigContent(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithData(c, info)
}

// @Tags SSH
// @Summary Update SSH configuration file content on host
// @Description Update SSH configuration file content on host
// @Accept json
// @Param host_id path uint true "Host ID"
// @Param content body string true "Content"
// @Success 200
// @Router /ssh/config/content/{host_id} [put]
func (s *SSHMan) UpdateSSHConfigContent(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host_id"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host_id", err)
		return
	}

	var content string
	if err := c.ShouldBindJSON(&content); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid content", err)
		return
	}

	req := model.ContentUpdate{
		HostID:  uint(hostID),
		Content: content,
	}

	if err := s.updateSSHContent(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags SSH
// @Summary Operate SSH
// @Description modify SSH service status on host
// @Accept json
// @Produce json
// @Param host_id path uint true "Host ID"
// @Param operation body string true "Operation, can be 'enable' or 'disable'"
// @Success 200 "No Content"
// @Router /ssh/operate/{host_id} [post]
func (s *SSHMan) OperateSSH(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host_id"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host_id", err)
		return
	}

	var operation string
	if err := c.ShouldBindJSON(&operation); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid operation", err)
		return
	}

	req := model.SSHOperate{
		HostID:    uint(hostID),
		Operation: operation,
	}

	if err := s.operateSSH(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags SSH
// @Summary Generate host SSH secret
// @Description Generate host SSH secret
// @Accept json
// @Produce json
// @Param host_id path uint true "Host ID"
// @Param request body model.GenerateKey true "request"
// @Success 200
// @Router /ssh/keys/{host_id} [post]
func (s *SSHMan) CreateKey(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host_id"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host_id", err)
		return
	}

	var req model.GenerateKey
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	req.HostID = uint(hostID)

	if err := s.createKey(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags SSH
// @Summary Get SSH secrets on host
// @Description Get SSH secrets on host
// @Accept json
// @Produce json
// @Param host_id path uint true "Host ID"
// @Param keyword query string false "Keyword"
// @Success 200
// @Router /ssh/keys/{host_id} [get]
func (s *SSHMan) ListKey(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host_id"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host_id", err)
		return
	}

	keyword := c.Query("keyword")

	req := model.ListKey{
		HostID:  uint(hostID),
		Keyword: keyword,
	}

	data, err := s.listKeys(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags SSH
// @Summary Get SSH logs on host
// @Description Get SSH logs on host
// @Accept json
// @Produce json
// @Param host_id path uint true "Host ID"
// @Param page query int true "Page number"
// @Param page_size query int true "Page size"
// @Param info query string false "Info"
// @Param status query string true "Status can be 'Success' or 'Failed' or 'All'"
// @Success 200 {object} model.SSHLog
// @Router /ssh/logs/{host_id} [get]
func (s *SSHMan) LoadSSHLogs(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host_id"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host_id", err)
		return
	}

	var req model.SearchSSHLog
	if err := helper.CheckQueryAndValidate(&req, c); err != nil {
		return
	}
	req.HostID = uint(hostID)

	data, err := s.loadLog(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithData(c, data)
}

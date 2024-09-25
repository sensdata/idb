package sshman

import (
	_ "embed"
	"fmt"
	"net/http"

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

	baseUrl := fmt.Sprintf("http://%s:%d", "0.0.0.0", conn.CONFMAN.GetConfig().Port)
	s.restyClient = resty.New().
		SetBaseURL(baseUrl).
		SetHeader("Content-Type", "application/json")

	api.API.SetUpPluginRouters(
		"ssh",
		[]plugin.PluginRoute{
			{Method: "GET", Path: "/info", Handler: s.GetPluginInfo},
			{Method: "GET", Path: "/menu", Handler: s.GetMenu},
			{Method: "POST", Path: "/config", Handler: s.GetSSHConfig},
			{Method: "POST", Path: "/config/update", Handler: s.UpdateSSHConfig},
			{Method: "POST", Path: "/config/content", Handler: s.GetSSHConfigContent},
			{Method: "POST", Path: "/config/content/update", Handler: s.UpdateSSHConfigContent},
			{Method: "POST", Path: "/operate", Handler: s.OperateSSH},
			{Method: "POST", Path: "/key/create", Handler: s.CreateKey},
			{Method: "POST", Path: "/key/list", Handler: s.ListKey},
			{Method: "POST", Path: "/log", Handler: s.LoadSSHLogs},
		},
	)
	global.LOG.Info("sshman init end")
}

func (s *SSHMan) Release() {

}

// @Tags SSH
// @Summary Plugin info
// @Description 插件信息
// @Accept json
// @Success 200 {array} plugin.PluginInfo
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
// @Summary Plugin menu
// @Description 插件菜单
// @Accept json
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
// @Summary Load host SSH setting info
// @Description 加载 SSH 配置信息
// @Param request body model.SSHConfigReq true "request"
// @Success 200 {object} model.SSHInfo
// @Router /ssh/config [post]
func (s *SSHMan) GetSSHConfig(c *gin.Context) {
	var req model.SSHConfigReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	info, err := s.getSSHConfig(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithData(c, info)
}

// @Tags SSH
// @Summary Update host SSH setting
// @Description 更新 SSH 配置
// @Accept json
// @Param request body model.SSHUpdate true "request"
// @Success 200
// @Router /ssh/config/update [post]
func (s *SSHMan) UpdateSSHConfig(c *gin.Context) {
	var req model.SSHUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := s.updateSSH(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags SSH
// @Summary Load host SSH setting file content
// @Description 加载 SSH 配置文件内容
// @Param request body model.SSHConfigReq true "request"
// @Success 200 {object} model.SSHConfigContent
// @Router /ssh/config/content [post]
func (s *SSHMan) GetSSHConfigContent(c *gin.Context) {
	var req model.SSHConfigReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	info, err := s.getSSHConfigContent(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithData(c, info)
}

// @Tags SSH
// @Summary Update host SSH setting
// @Description 更新 SSH 配置文件内容
// @Accept json
// @Param request body model.ContentUpdate true "request"
// @Success 200
// @Router /ssh/config/content/update [post]
func (s *SSHMan) UpdateSSHConfigContent(c *gin.Context) {
	var req model.ContentUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := s.updateSSHContent(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags SSH
// @Summary Operate SSH
// @Description 修改 SSH 服务状态
// @Accept json
// @Param request body model.SSHOperate true "request"
// @Success 200
// @Router /ssh/operate [post]
func (s *SSHMan) OperateSSH(c *gin.Context) {
	var req model.SSHOperate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := s.operateSSH(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags SSH
// @Summary Generate host SSH secret
// @Description 生成 SSH 密钥
// @Accept json
// @Param request body model.GenerateKey true "request"
// @Success 200
// @Router /ssh/key/create [post]
func (s *SSHMan) CreateKey(c *gin.Context) {
	var req model.GenerateKey
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := s.createKey(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags SSH
// @Summary Load host SSH secret
// @Description 枚举 SSH 密钥
// @Accept json
// @Param request body model.ListKey true "request"
// @Success 200
// @Router /ssh/key/list [post]
func (s *SSHMan) ListKey(c *gin.Context) {
	var req model.ListKey
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	data, err := s.listKeys(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags SSH
// @Summary Load host SSH logs
// @Description 获取 SSH 登录日志
// @Accept json
// @Param request body model.SearchSSHLog true "request"
// @Success 200 {object} model.SSHLog
// @Router /ssh/log [post]
func (s *SSHMan) LoadSSHLogs(c *gin.Context) {
	var req model.SearchSSHLog
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	data, err := s.loadLog(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithData(c, data)
}

package sshman

import (
	_ "embed"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/core/api"
	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/helper"
	"github.com/sensdata/idb/core/plugin"
	"gopkg.in/yaml.v2"
)

const sshPath = "/etc/ssh/sshd_config"

type SSHMan struct {
	config    plugin.PluginConfig
	cmdHelper *helper.CmdHelper
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

	// TODO: 根据配置传入
	s.cmdHelper = helper.NewCmdHelper("127.0.0.1", "8080", nil)

	api.API.SetUpPluginRouters(
		"ssh",
		[]plugin.PluginRoute{
			{Method: "GET", Path: "/info", Handler: s.GetPluginInfo},
			{Method: "GET", Path: "/menu", Handler: s.GetMenu},
			{Method: "GET", Path: "/config", Handler: s.GetSSHConfig},
		},
	)
	global.LOG.Info("sshman init end")
}

func (s *SSHMan) Release() {

}

// GetPluginInfo 处理 /fileman/plugin_info 请求
func (s *SSHMan) GetPluginInfo(c *gin.Context) {
	pluginInfo, err := s.getPluginInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"type": "info", "payload": pluginInfo})
}

// GetMenu 处理 /fileman/menu 请求
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
// @Success 200 {object} model.SSHInfo
// @Router /ssh/config [post]
func (s *SSHMan) GetSSHConfig(c *gin.Context) {
	info, err := s.getSSHConfig()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithData(c, info)
}

package systemctl

import (
	_ "embed"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/core/api"
	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/helper"
	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/plugin"
	"gopkg.in/yaml.v2"
)

type SystemCtl struct {
	config    plugin.PluginConfig
	cmdHelper *helper.CmdHelper
}

var Plugin = SystemCtl{}

//go:embed plug.yaml
var plugYAML []byte

func (s *SystemCtl) Initialize() {
	global.LOG.Info("systemctl init begin")

	if err := yaml.Unmarshal(plugYAML, &s.config); err != nil {
		global.LOG.Error("Failed to load fileman yaml: %v", err)
		return
	}

	// TODO: 根据配置传入
	s.cmdHelper = helper.NewCmdHelper("127.0.0.1", "8080", nil)

	api.API.SetUpPluginRouters(
		"sysctl",
		[]plugin.PluginRoute{
			{Method: "GET", Path: "/info", Handler: s.GetPluginInfo},
			{Method: "GET", Path: "/menu", Handler: s.GetMenu},
			{Method: "GET", Path: "/boot", Handler: s.ServiceBoot},
		},
	)

	global.LOG.Info("systemctl init end")
}

func (s *SystemCtl) Release() {}

// @Tags Sysctl
// @Summary Plugin info
// @Description 插件信息
// @Accept json
// @Success 200 {array} plugin.PluginInfo
// @Router /sysctl/info [get]
func (s *SystemCtl) GetPluginInfo(c *gin.Context) {
	pluginInfo, err := s.getPluginInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"type": "info", "payload": pluginInfo})
}

// @Tags Sysctl
// @Summary Plugin menu
// @Description 插件菜单
// @Accept json
// @Success 200 {array} plugin.PluginInfo
// @Router /sysctl/menu [get]
func (s *SystemCtl) GetMenu(c *gin.Context) {
	menuItems, err := s.getMenus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"type": "menu", "payload": menuItems})
}

func (s *SystemCtl) getPluginInfo() (plugin.PluginInfo, error) {
	return s.config.Plugin, nil
}

func (s *SystemCtl) getMenus() ([]plugin.MenuItem, error) {
	return s.config.Menu, nil
}

// @Tags SystemCtl
// @Summary run systemctl start/stop/restart service commands
// @Description 运行systemctl的服务启动相关命令
// @Success 200
// @Router /sysctl/boot [post]
func (s *SystemCtl) ServiceBoot(c *gin.Context) {
	var req model.ServiceBootReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	err := s.serviceBoot(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithData(c, nil)
}

func (s *SystemCtl) serviceBoot(req model.ServiceBootReq) error {
	_, err := s.cmdHelper.RunSystemCtl(req.HostID, req.ServiceName)
	if err != nil {
		global.LOG.Info("service boot err %v", err)
		return err
	}
	return nil
}

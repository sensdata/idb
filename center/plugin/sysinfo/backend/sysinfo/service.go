package sysinfo

import (
	"net/http"

	_ "embed"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/sensdata/idb/center/core/api"
	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/model"
	"gopkg.in/yaml.v2"

	"github.com/sensdata/idb/core/plugin"
)

type SysInfo struct {
	config      plugin.PluginConfig
	restyClient *resty.Client
}

var Plugin = SysInfo{}

//go:embed plug.yaml
var plugYAML []byte

func (s *SysInfo) Initialize() {
	global.LOG.Info("sysinfo init begin")

	if err := yaml.Unmarshal(plugYAML, &s.config); err != nil {
		global.LOG.Error("Failed to load sysinfo yaml: %v", err)
		return
	}

	s.restyClient = resty.New().
		SetBaseURL("http://127.0.0.1:8080").
		SetHeader("Content-Type", "application/json")

	api.API.SetUpPluginRouters(
		"sysinfo",
		[]plugin.PluginRoute{
			{Method: "GET", Path: "/info", Handler: s.GetPluginInfo},
			{Method: "GET", Path: "/menu", Handler: s.GetMenu},
			{Method: "GET", Path: "/overview", Handler: s.GetOverview},
			{Method: "GET", Path: "/network", Handler: s.GetNetwork},
			{Method: "GET", Path: "/system", Handler: s.GetSystemInfo},
			{Method: "GET", Path: "/config", Handler: s.GetConfig},
			{Method: "GET", Path: "/hardware", Handler: s.GetHardware},
		},
	)

	global.LOG.Info("sysinfo init end")
}

func (s *SysInfo) Release() {

}

// @Tags Sysinfo
// @Summary Plugin info
// @Description 插件信息
// @Accept json
// @Success 200 {array} plugin.PluginInfo
// @Router /sysinfo/info [get]
func (s *SysInfo) GetPluginInfo(c *gin.Context) {
	pluginInfo, err := s.getPluginInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"type": "info", "payload": pluginInfo})
}

// @Tags Sysinfo
// @Summary Plugin menu
// @Description 插件菜单
// @Accept json
// @Success 200 {array} plugin.PluginInfo
// @Router /sysinfo/menu [get]
func (s *SysInfo) GetMenu(c *gin.Context) {
	menuItems, err := s.getMenus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"type": "menu", "payload": menuItems})
}

// @Tags Sysinfo
// @Summary Get overview
// @Description 获取系统概览
// @Accept json
// @Success 200 {array} model.Overview
// @Router /sysinfo/overview [get]
func (s *SysInfo) GetOverview(c *gin.Context) {
	overview, err := s.getOverview()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"type": "overview", "payload": overview})
}

// @Tags Sysinfo
// @Summary Get network
// @Description 获取网络信息
// @Accept json
// @Success 200 {array} model.NetworkInfo
// @Router /sysinfo/network [get]
func (s *SysInfo) GetNetwork(c *gin.Context) {
	network, err := s.getNetwork()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"type": "network", "payload": network})
}

// @Tags Sysinfo
// @Summary Get system info
// @Description 获取系统信息
// @Accept json
// @Success 200 {array} model.SystemInfo
// @Router /sysinfo/system [get]
func (s *SysInfo) GetSystemInfo(c *gin.Context) {
	system, err := s.getSystemInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"type": "network", "payload": system})
}

// @Tags Sysinfo
// @Summary Get system config
// @Description 获取系统配置
// @Accept json
// @Success 200 {array} model.SystemConfig
// @Router /sysinfo/config [get]
func (s *SysInfo) GetConfig(c *gin.Context) {
	config, err := s.getConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"type": "network", "payload": config})
}

// @Tags Sysinfo
// @Summary Get hardware info
// @Description 获取硬件信息
// @Accept json
// @Success 200 {array} model.HardwareInfo
// @Router /sysinfo/hardware [get]
func (s *SysInfo) GetHardware(c *gin.Context) {
	hardware, err := s.getHardware()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"type": "network", "payload": hardware})
}

func (s *SysInfo) getPluginInfo() (plugin.PluginInfo, error) {
	return s.config.Plugin, nil
}

func (s *SysInfo) getMenus() ([]plugin.MenuItem, error) {
	return s.config.Menu, nil
}

func (s *SysInfo) getConfig() (model.SystemConfig, error) {
	return model.SystemConfig{}, nil
}

func (s *SysInfo) getHardware() (model.HardwareInfo, error) {
	return model.HardwareInfo{}, nil
}

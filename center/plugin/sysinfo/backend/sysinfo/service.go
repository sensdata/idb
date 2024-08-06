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

// GetPluginInfo 处理 /sysinfo/plugin_info 请求
func (s *SysInfo) GetPluginInfo(c *gin.Context) {
	pluginInfo, err := s.getPluginInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"type": "info", "payload": pluginInfo})
}

// GetMenu 处理 /sysinfo/menu 请求
func (s *SysInfo) GetMenu(c *gin.Context) {
	menuItems, err := s.getMenus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"type": "menu", "payload": menuItems})
}

// GetOverview 处理 /sysinfo/overview 请求
func (s *SysInfo) GetOverview(c *gin.Context) {
	overview, err := s.getOverview()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"type": "overview", "payload": overview})
}

func (s *SysInfo) GetNetwork(c *gin.Context) {
	network, err := s.getNetwork()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"type": "network", "payload": network})
}

func (s *SysInfo) GetSystemInfo(c *gin.Context) {
	system, err := s.getSystemInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"type": "network", "payload": system})
}

func (s *SysInfo) GetConfig(c *gin.Context) {
	config, err := s.getConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"type": "network", "payload": config})
}

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

func (s *SysInfo) getNetwork() (model.NetworkInfo, error) {
	return model.NetworkInfo{}, nil
}

func (s *SysInfo) getSystemInfo() (model.SystemInfo, error) {
	return model.SystemInfo{}, nil
}

func (s *SysInfo) getConfig() (model.SystemConfig, error) {
	return model.SystemConfig{}, nil
}

func (s *SysInfo) getHardware() (model.HardwareInfo, error) {
	return model.HardwareInfo{}, nil
}

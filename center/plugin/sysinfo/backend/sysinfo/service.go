package sysinfo

import (
	"fmt"
	"net/http"

	_ "embed"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/sensdata/idb/center/core/api"
	"github.com/sensdata/idb/center/core/conn"
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

	baseUrl := fmt.Sprintf("http://%s:%d", "0.0.0.0", conn.CONFMAN.GetConfig().Port)
	s.restyClient = resty.New().
		SetBaseURL(baseUrl).
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
// @Summary Get plugin info
// @Description Get plugin information
// @Accept json
// @Produce json
// @Success 200 {object} plugin.PluginInfo
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
// @Summary Get plugin menu
// @Description Get plugin menu items
// @Accept json
// @Produce json
// @Success 200 {array} plugin.MenuItem
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
// @Summary Get system overview info
// @Description Get system overview info
// @Accept json
// @Produce json
// @Success 200 {object} model.Overview
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
// @Summary Get network info
// @Description Get network info
// @Accept json
// @Produce json
// @Success 200 {object} model.NetworkInfo
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
// @Description Get system info
// @Accept json
// @Produce json
// @Success 200 {object} model.SystemInfo
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
// @Description Get system config
// @Accept json
// @Produce json
// @Success 200 {object} model.SystemConfig
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
// @Description Get hardware info
// @Accept json
// @Produce json
// @Success 200 {object} model.HardwareInfo
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

func (s *SysInfo) sendAction(actionRequest model.HostAction) (*model.ActionResponse, error) {
	var actionResponse model.ActionResponse

	resp, err := s.restyClient.R().
		SetBody(actionRequest).
		SetResult(&actionResponse).
		Post("/actions") // 修改URL路径

	if err != nil {
		global.LOG.Error("failed to send request: %v", err)
		return nil, fmt.Errorf("failed to send request: %v", err)
	}

	if resp.StatusCode() != 200 {
		global.LOG.Error("received error response: %s", resp.Status())
		return nil, fmt.Errorf("received error response: %s", resp.Status())
	}

	return &actionResponse, nil
}

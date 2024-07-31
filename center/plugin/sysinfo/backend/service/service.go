package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/core/api"
	"github.com/sensdata/idb/center/plugin/sysinfo/backend/model"

	"github.com/sensdata/idb/core/plugin"
	"github.com/sensdata/idb/core/utils"
)

var config plugin.PluginConfig

func init() {
	err := utils.LoadYaml("plug.yaml", &config)
	if err != nil {
		panic(err)
	}

	api.API.SetUpPluginRouters(
		"sysinfo",
		[]plugin.PluginRoute{
			{Method: "GET", Path: "/info", Handler: GetPluginInfo},
			{Method: "GET", Path: "/menu", Handler: GetMenu},
			{Method: "GET", Path: "/overview", Handler: GetOverview},
			{Method: "GET", Path: "/network", Handler: GetNetwork},
			{Method: "GET", Path: "/system", Handler: GetSystemInfo},
			{Method: "GET", Path: "/config", Handler: GetConfig},
			{Method: "GET", Path: "/hardware", Handler: GetHardware},
		},
	)
}

// GetPluginInfo 处理 /sysinfo/plugin_info 请求
func GetPluginInfo(c *gin.Context) {
	pluginInfo, err := getPluginInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"type": "info", "payload": pluginInfo})
}

// GetMenu 处理 /sysinfo/menu 请求
func GetMenu(c *gin.Context) {
	menuItems, err := getMenus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"type": "menu", "payload": menuItems})
}

// GetOverview 处理 /sysinfo/summary 请求
func GetOverview(c *gin.Context) {
	overview, err := getOverview()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"type": "overview", "payload": overview})
}

func GetNetwork(c *gin.Context) {
	network, err := getNetwork()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"type": "network", "payload": network})
}

func GetSystemInfo(c *gin.Context) {
	system, err := getSystemInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"type": "network", "payload": system})
}

func GetConfig(c *gin.Context) {
	config, err := getConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"type": "network", "payload": config})
}

func GetHardware(c *gin.Context) {
	hardware, err := getHardware()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"type": "network", "payload": hardware})
}

func getPluginInfo() (plugin.PluginInfo, error) {
	return config.Plugin, nil
}

func getMenus() ([]plugin.MenuItem, error) {
	return config.Menu, nil
}

func getOverview() (model.Overview, error) {
	return model.Overview{}, nil
}

func getNetwork() (model.NetworkInfo, error) {
	return model.NetworkInfo{}, nil
}

func getSystemInfo() (model.SystemInfo, error) {
	return model.SystemInfo{}, nil
}

func getConfig() (model.SystemConfig, error) {
	return model.SystemConfig{}, nil
}

func getHardware() (model.HardwareInfo, error) {
	return model.HardwareInfo{}, nil
}

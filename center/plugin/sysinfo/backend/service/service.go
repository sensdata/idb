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
			{Method: "GET", Path: "/summary", Handler: GetSummary},
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

// GetSummary 处理 /sysinfo/summary 请求
func GetSummary(c *gin.Context) {
	summary, err := getSummary()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"type": "summary", "payload": summary})
}

func getPluginInfo() (plugin.PluginInfo, error) {
	return config.Plugin, nil
}

func getMenus() ([]plugin.MenuItem, error) {
	return config.Menu, nil
}

func getSummary() (model.Overview, error) {
	return model.Overview{}, nil
}

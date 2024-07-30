package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/core/api"

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
			{Method: "GET", Path: "/menu", Handler: HandleMenu},
			{Method: "GET", Path: "/plugin_info", Handler: HandlePluginInfo},
		},
	)
}

// HandleMenu 处理 /sysinfo/menu 请求
func HandleMenu(c *gin.Context) {
	menuItems, err := getMenus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"type": "menuItems", "payload": menuItems})
}

// HandlePluginInfo 处理 /sysinfo/plugin_info 请求
func HandlePluginInfo(c *gin.Context) {
	pluginInfo, err := getPluginInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"type": "pluginInfo", "payload": pluginInfo})
}

func getPluginInfo() (plugin.PluginInfo, error) {
	return config.Plugin, nil
}

func getMenus() ([]plugin.MenuItem, error) {
	return config.Menu, nil
}

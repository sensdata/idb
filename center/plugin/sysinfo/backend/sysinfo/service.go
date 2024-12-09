package sysinfo

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	_ "embed"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/sensdata/idb/center/core/api"
	"github.com/sensdata/idb/center/core/conn"
	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/helper"
	"github.com/sensdata/idb/core/log"
	"github.com/sensdata/idb/core/model"
	"gopkg.in/yaml.v2"

	"github.com/sensdata/idb/core/plugin"
)

type SysInfo struct {
	plugin      plugin.Plugin
	pluginConf  plugin.PluginConf
	restyClient *resty.Client
}

var LOG *log.Log

//go:embed plug.yaml
var plugYAML []byte

//go:embed conf.yaml
var confYAML []byte

func (s *SysInfo) Initialize() {
	global.LOG.Info("sysinfo init begin")

	if err := yaml.Unmarshal(plugYAML, &s.plugin); err != nil {
		global.LOG.Error("Failed to load sysinfo yaml: %v", err)
		return
	}

	confPath := filepath.Join(constant.CenterConfDir, "files", "conf.yaml")
	// 检查配置文件的目录是否存在
	if err := os.MkdirAll(filepath.Dir(confPath), os.ModePerm); err != nil {
		global.LOG.Error("Failed to create conf directory: %v \n", err)
		return
	}
	// 检查配置文件是否存在
	if _, err := os.Stat(confPath); os.IsNotExist(err) {
		// 创建配置文件并写入默认内容
		if err := os.WriteFile(confPath, confYAML, 0644); err != nil {
			global.LOG.Error("Failed to create conf: %v \n", err)
			return
		}
	}
	// 读取文件内容
	data, err := os.ReadFile(confPath)
	if err != nil {
		global.LOG.Error("Failed to read conf: %v \n", err)
		return
	}
	// 解析 YAML 内容
	if err := yaml.Unmarshal(data, &s.pluginConf); err != nil {
		global.LOG.Error("Failed to load conf: %v", err)
		return
	}

	//初始化日志模块
	if LOG == nil {
		logger, err := log.InitLogger(s.pluginConf.Items.LogDir, "files.log")
		if err != nil {
			global.LOG.Error("Failed to initialize logger: %v \n", err)
			return
		}
		LOG = logger
	}

	baseUrl := fmt.Sprintf("http://%s:%d/api/v1", "127.0.0.1", conn.CONFMAN.GetConfig().Port)
	s.restyClient = resty.New().
		SetBaseURL(baseUrl).
		SetHeader("Content-Type", "application/json")

	api.API.SetUpPluginRouters(
		"sysinfo",
		[]plugin.PluginRoute{
			{Method: "GET", Path: "/info", Handler: s.GetPluginInfo},
			{Method: "GET", Path: "/menu", Handler: s.GetMenu},
			{Method: "GET", Path: "/:host/overview", Handler: s.GetOverview},
			{Method: "GET", Path: "/:host/network", Handler: s.GetNetwork},
			{Method: "GET", Path: "/:host/system", Handler: s.GetSystemInfo},
			{Method: "GET", Path: "/:host/config", Handler: s.GetConfig},
			{Method: "GET", Path: "/:host/hardware", Handler: s.GetHardware},
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
// @Param host path uint true "Host ID"
// @Success 200 {object} model.Overview
// @Router /sysinfo/{host}/overview [get]
func (s *SysInfo) GetOverview(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	overview, err := s.getOverview(uint(hostID))
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
// @Param host path uint true "Host ID"
// @Success 200 {object} model.NetworkInfo
// @Router /sysinfo/{host}/network [get]
func (s *SysInfo) GetNetwork(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	network, err := s.getNetwork(uint(hostID))
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
// @Param host path uint true "Host ID"
// @Success 200 {object} model.SystemInfo
// @Router /sysinfo/{host}/system [get]
func (s *SysInfo) GetSystemInfo(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	system, err := s.getSystemInfo(uint(hostID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"type": "network", "payload": system})
}

// @Tags Sysinfo
// @Summary Get system config
// @Description (not implemented yet) Get system config
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Success 200 {object} model.SystemConfig
// @Router /sysinfo/{host}/config [get]
// @Deprecated
func (s *SysInfo) GetConfig(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	config, err := s.getConfig(uint(hostID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"type": "network", "payload": config})
}

// @Tags Sysinfo
// @Summary Get hardware info
// @Description (not implemented yet) Get hardware info
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Success 200 {object} model.HardwareInfo
// @Router /sysinfo/{host}/hardware [get]
// @Deprecated
func (s *SysInfo) GetHardware(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	hardware, err := s.getHardware(uint(hostID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"type": "network", "payload": hardware})
}

func (s *SysInfo) getPluginInfo() (plugin.PluginInfo, error) {
	return s.plugin.Info, nil
}

func (s *SysInfo) getMenus() ([]plugin.MenuItem, error) {
	return s.plugin.Menu, nil
}

func (s *SysInfo) getConfig(_ uint) (model.SystemConfig, error) {
	return model.SystemConfig{}, nil
}

func (s *SysInfo) getHardware(_ uint) (model.HardwareInfo, error) {
	return model.HardwareInfo{}, nil
}

func (s *SysInfo) sendAction(actionRequest model.HostAction) (*model.ActionResponse, error) {
	var actionResponse model.ActionResponse

	resp, err := s.restyClient.R().
		SetBody(actionRequest).
		SetResult(&actionResponse).
		Post("/actions") // 修改URL路径

	if err != nil {
		LOG.Error("failed to send request: %v", err)
		return nil, fmt.Errorf("failed to send request: %v", err)
	}

	if resp.StatusCode() != 200 {
		LOG.Error("received error response: %s", resp.Status())
		return nil, fmt.Errorf("received error response: %s", resp.Status())
	}

	return &actionResponse, nil
}

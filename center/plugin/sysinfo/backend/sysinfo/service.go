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
	"github.com/sensdata/idb/center/core/api/service"
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

	settingService := service.NewISettingsService()
	settingInfo, _ := settingService.Settings()
	scheme := "http"
	if settingInfo.Https == "yes" {
		scheme = "https"
	}
	host := global.Host
	if settingInfo.BindDomain != "" && settingInfo.BindDomain != host {
		host = settingInfo.BindDomain
	}
	baseUrl := fmt.Sprintf("%s://%s:%d/api/v1", scheme, host, settingInfo.BindPort)

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
			{Method: "GET", Path: "/:host/hardware", Handler: s.GetHardware},
			{Method: "GET", Path: "/:host/settings", Handler: s.GetSysSettings},
			{Method: "POST", Path: "/:host/action/upd/time", Handler: s.SetTime},
			{Method: "POST", Path: "/:host/action/upd/timezone", Handler: s.SetTimeZone},
			{Method: "POST", Path: "/:host/action/sync/time", Handler: s.SyncTime},
			{Method: "POST", Path: "/:host/action/memcache/clear", Handler: s.ClearMemCache},
			{Method: "POST", Path: "/:host/action/memcache/auto/set", Handler: s.SetAutoClearInterval},
			{Method: "POST", Path: "/:host/action/swap/create", Handler: s.CreateSwap},
			{Method: "POST", Path: "/:host/action/swap/delete", Handler: s.DeleteSwap},
			{Method: "POST", Path: "/:host/action/upd/dns", Handler: s.UpdateDnsSettings},
			{Method: "POST", Path: "/:host/action/upd/settings", Handler: s.UpdateSysSetting},
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

func (s *SysInfo) getPluginInfo() (plugin.PluginInfo, error) {
	return s.plugin.Info, nil
}

func (s *SysInfo) getMenus() ([]plugin.MenuItem, error) {
	return s.plugin.Menu, nil
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
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithData(c, overview)
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
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithData(c, network)
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
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithData(c, system)
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
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithData(c, hardware)
}

func (s *SysInfo) getHardware(_ uint) (model.HardwareInfo, error) {
	return model.HardwareInfo{}, nil
}

// @Tags Sysinfo
// @Summary Get system settings
// @Description Get system settings of the specified host
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Success 200 {object} model.SystemSettings
// @Router /sysinfo/{host}/settings [get]
func (s *SysInfo) GetSysSettings(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	settings, err := s.getSystemSettings(uint(hostID))
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithData(c, settings)
}

// @Tags Sysinfo
// @Summary Set system time
// @Description Set system time for the specified host
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param request body model.SetTimeReq true "Time settings"
// @Success 200
// @Router /sysinfo/{host}/action/upd/time [post]
func (s *SysInfo) SetTime(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	var req model.SetTimeReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid params", err)
		return
	}

	err = s.setTime(uint(hostID), req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, "Failed to set time", err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Sysinfo
// @Summary Set system timezone
// @Description Set system timezone for the specified host
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param request body model.SetTimezoneReq true "Time-zone settings"
// @Success 200
// @Router /sysinfo/{host}/action/upd/timezone [post]
func (s *SysInfo) SetTimeZone(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	var req model.SetTimezoneReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid params", err)
		return
	}

	err = s.setTimeZone(uint(hostID), req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, "Failed to set time-zone", err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Sysinfo
// @Summary Sync system time
// @Description Sync system time with NTP server for the specified host
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Success 200
// @Router /sysinfo/{host}/action/sync/time [post]
func (s *SysInfo) SyncTime(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	err = s.syncTime(uint(hostID))
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, "Failed to sync time", err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Sysinfo
// @Summary Clear memory cache
// @Description Clear memory cache for the specified host
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Success 200
// @Router /sysinfo/{host}/action/memcache/clear [post]
func (s *SysInfo) ClearMemCache(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	err = s.clearMemCache(uint(hostID))
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, "Failed to clear cache", err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Sysinfo
// @Summary Set auto clear interval
// @Description Set auto clear interval for the specified host
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param request body model.AutoClearMemCacheReq true "Auto clear interval settings"
// @Success 200
// @Router /sysinfo/{host}/action/memcache/auto/set [post]
func (s *SysInfo) SetAutoClearInterval(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	var req model.AutoClearMemCacheReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid params", err)
		return
	}

	err = s.setAutoClearInterval(uint(hostID), req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, "Failed to set auto clear interval", err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Sysinfo
// @Summary Create swap
// @Description Create swap for the specified host
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param request body model.CreateSwapReq true "Create swap settings"
// @Success 200
// @Router /sysinfo/{host}/action/swap/create [post]
func (s *SysInfo) CreateSwap(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	var req model.CreateSwapReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid params", err)
		return
	}

	err = s.createSwap(uint(hostID), req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, "Failed to creat swat", err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Sysinfo
// @Summary Delete swap
// @Description Delete swap for the specified host
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Success 200
// @Router /sysinfo/{host}/action/swap/delete [post]
func (s *SysInfo) DeleteSwap(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	err = s.deleteSwap(uint(hostID))
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, "Failed to creat swat", err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Sysinfo
// @Summary Update DNS settings
// @Description Update DNS settings for the specified host
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param request body model.UpdateDnsSettingsReq true "DNS settings"
// @Success 200
// @Router /sysinfo/{host}/action/upd/dns [post]
func (s *SysInfo) UpdateDnsSettings(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	var req model.UpdateDnsSettingsReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid params", err)
		return
	}
	err = s.updateDNS(uint(hostID), req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, "Failed to update dns settings", err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Sysinfo
// @Summary Update system settings
// @Description Update system settings for the specified host
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param request body model.UpdateSystemSettingsReq true "System settings"
// @Success 200
// @Router /sysinfo/{host}/action/upd/settings [post]
func (s *SysInfo) UpdateSysSetting(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	var req model.UpdateSystemSettingsReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid params", err)
		return
	}
	err = s.updateSystemSettings(uint(hostID), req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, "Failed to update system settings", err)
		return
	}
	helper.SuccessWithData(c, nil)
}

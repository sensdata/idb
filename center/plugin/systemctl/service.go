package systemctl

import (
	"crypto/tls"
	_ "embed"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/sensdata/idb/center/core/api"
	"github.com/sensdata/idb/center/core/api/service"
	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/helper"
	"github.com/sensdata/idb/core/log"
	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/plugin"
	"gopkg.in/yaml.v2"
)

type SystemCtl struct {
	plugin      plugin.Plugin
	pluginConf  plugin.PluginConf
	restyClient *resty.Client
}

var LOG *log.Log

//go:embed plug.yaml
var plugYAML []byte

//go:embed conf.yaml
var confYAML []byte

func (s *SystemCtl) Initialize() {
	global.LOG.Info("systemctl init begin")

	if err := yaml.Unmarshal(plugYAML, &s.plugin); err != nil {
		global.LOG.Error("Failed to load sysctl yaml: %v", err)
		return
	}

	confPath := filepath.Join(constant.CenterConfDir, "sysctl", "conf.yaml")
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

	if settingInfo.Https == "yes" {
		// 创建 TLS 配置
		cert, err := tls.X509KeyPair(global.CertPem, global.KeyPem)
		if err != nil {
			global.LOG.Error("Failed to create cert: %v", err)
			return
		}

		tlsConfig := &tls.Config{
			Certificates:       []tls.Certificate{cert}, // 设置服务器证书
			MinVersion:         tls.VersionTLS13,        // 设置最小 TLS 版本
			InsecureSkipVerify: true,
		}
		s.restyClient.SetTLSClientConfig(tlsConfig)
	}

	api.API.SetUpPluginRouters(
		"sysctl",
		[]plugin.PluginRoute{
			{Method: "GET", Path: "/info", Handler: s.GetPluginInfo},
			{Method: "GET", Path: "/menu", Handler: s.GetMenu},
			{Method: "GET", Path: "/:host/operate", Handler: s.OperateService},
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
	return s.plugin.Info, nil
}

func (s *SystemCtl) getMenus() ([]plugin.MenuItem, error) {
	return s.plugin.Menu, nil
}

// @Tags Sysctl
// @Summary Operate service
// @Description Performing operations on a systemd-managed service on the specified host. Supported actions include enabling/disabling on boot, starting, stopping, restarting, reloading, and get status of the service.
// @Param host path uint true "Host ID"
// @Param request body model.OperateServiceReq true "Service operation"
// @Success 200 {object} model.ServiceOperateResult
// @Router /sysctl/{host}/operate [post]
func (s *SystemCtl) OperateService(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	var req model.OperateServiceReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	result, err := s.serviceOperate(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}
	helper.SuccessWithData(c, result)
}

func (s *SystemCtl) serviceOperate(hostID uint64, req model.OperateServiceReq) (*model.ServiceOperateResult, error) {
	var result model.ServiceOperateResult

	var command string
	switch req.Operation {
	case "start", "stop", "restart", "reload", "enable", "disable":
		command = fmt.Sprintf("systemctl %s %s", req.Operation, req.Service)
	case "status":
		command = fmt.Sprintf("systemctl show %s --property=ActiveState,SubState", req.Service)
	default:
		return &result, errors.New("unsupported operation")
	}
	commandResult, err := s.sendCommand(uint(hostID), command)
	if err != nil {
		LOG.Error("Failed to send service operation command")
		return &result, err
	}
	LOG.Info("Service operation command result: %s", commandResult.Result)
	result.Result = commandResult.Result
	return &result, nil
}

func (s *SystemCtl) sendCommand(hostId uint, command string) (*model.CommandResult, error) {
	var commandResult model.CommandResult

	commandRequest := model.Command{
		HostID:  hostId,
		Command: command,
	}

	var commandResponse model.CommandResponse

	resp, err := s.restyClient.R().
		SetBody(commandRequest).
		SetResult(&commandResponse).
		Post("/commands")

	if err != nil {
		LOG.Error("failed to send request: %v", err)
		return &commandResult, fmt.Errorf("failed to send request: %v", err)
	}

	if resp.StatusCode() != 200 {
		LOG.Error("failed to send request: %v", err)
		return &commandResult, fmt.Errorf("received error response: %s", resp.Status())
	}

	LOG.Info("cmd response: %v", commandResponse)

	commandResult = commandResponse.Data

	return &commandResult, nil
}

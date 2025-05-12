package sshman

import (
	"crypto/tls"
	_ "embed"
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

type SSHMan struct {
	plugin      plugin.Plugin
	pluginConf  plugin.PluginConf
	restyClient *resty.Client
}

var LOG *log.Log

//go:embed plug.yaml
var plugYAML []byte

//go:embed conf.yaml
var confYAML []byte

func (s *SSHMan) Initialize() {
	global.LOG.Info("sshman init begin")

	if err := yaml.Unmarshal(plugYAML, &s.plugin); err != nil {
		global.LOG.Error("Failed to load fileman yaml: %v", err)
		return
	}

	confPath := filepath.Join(constant.CenterConfDir, "ssh", "conf.yaml")
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
		"ssh",
		[]plugin.PluginRoute{
			{Method: "GET", Path: "/info", Handler: s.GetPluginInfo},
			{Method: "GET", Path: "/menu", Handler: s.GetMenu},
			{Method: "GET", Path: "/:host/config", Handler: s.GetSSHConfig},
			{Method: "PUT", Path: "/:host/config", Handler: s.UpdateSSHConfig},
			{Method: "GET", Path: "/:host/config/content", Handler: s.GetSSHConfigContent},
			{Method: "PUT", Path: "/:host/config/content", Handler: s.UpdateSSHConfigContent},
			{Method: "POST", Path: "/:host/operate", Handler: s.OperateSSH},
			{Method: "POST", Path: "/:host/keys", Handler: s.CreateKey},
			{Method: "POST", Path: "/:host/keys/enable", Handler: s.EnableKey},
			{Method: "POST", Path: "/:host/keys/remove", Handler: s.RemoveKey},
			{Method: "GET", Path: "/:host/keys/download", Handler: s.DownloadPrivateKey},
			{Method: "POST", Path: "/:host/keys/password/set", Handler: s.SetKeyPassword},
			{Method: "POST", Path: "/:host/keys/password/update", Handler: s.UpdateKeyPassword},
			{Method: "POST", Path: "/:host/keys/password/clear", Handler: s.ClearKeyPassword},

			{Method: "GET", Path: "/:host/keys", Handler: s.ListKey},
			{Method: "GET", Path: "/:host/logs", Handler: s.LoadSSHLogs},
		},
	)
	global.LOG.Info("sshman init end")
}

func (s *SSHMan) Release() {

}

// @Tags SSH
// @Summary Get plugin info
// @Description Get plugin information
// @Accept json
// @Produce json
// @Success 200 {object} plugin.PluginInfo
// @Router /ssh/info [get]
func (s *SSHMan) GetPluginInfo(c *gin.Context) {
	pluginInfo, err := s.getPluginInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"type": "info", "payload": pluginInfo})
}

// @Tags SSH
// @Summary Get plugin menu
// @Description Get plugin menu items
// @Accept json
// @Produce json
// @Success 200 {array} plugin.MenuItem
// @Router /ssh/menu [get]
func (s *SSHMan) GetMenu(c *gin.Context) {
	menuItems, err := s.getMenus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"type": "menu", "payload": menuItems})
}

func (s *SSHMan) getPluginInfo() (plugin.PluginInfo, error) {
	return s.plugin.Info, nil
}

func (s *SSHMan) getMenus() ([]plugin.MenuItem, error) {
	return s.plugin.Menu, nil
}

// @Tags SSH
// @Summary Get SSH configurations on host
// @Description Get SSH configurations on host
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Success 200 {object} model.SSHInfo
// @Router /ssh/{host}/config [get]
func (s *SSHMan) GetSSHConfig(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	info, err := s.getSSHConfig(hostID)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithData(c, info)
}

// @Tags SSH
// @Summary Update SSH configurations on host
// @Description Update SSH configurations on host
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param request body model.SSHUpdate true "request"
// @Success 200
// @Router /ssh/{host}/config [put]
func (s *SSHMan) UpdateSSHConfig(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	var req model.SSHUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := s.updateSSH(hostID, req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags SSH
// @Summary Get SSH config file content on host
// @Description Get SSH config file content on host
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Success 200 {object} model.SSHConfigContent
// @Router /ssh/{host}/config/content [get]
func (s *SSHMan) GetSSHConfigContent(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	info, err := s.getSSHConfigContent(hostID)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithData(c, info)
}

// @Tags SSH
// @Summary Update SSH configuration file content on host
// @Description Update SSH configuration file content on host
// @Accept json
// @Param host path uint true "Host ID"
// @Param content body string true "Content"
// @Success 200
// @Router /ssh/{host}/config/content [put]
func (s *SSHMan) UpdateSSHConfigContent(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	var req model.ContentUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := s.updateSSHContent(hostID, req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags SSH
// @Summary Operate SSH
// @Description modify SSH service status on host
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param request body model.SSHOperate true "Operation request"
// @Success 200 "No Content"
// @Router /ssh/{host}/operate [post]
func (s *SSHMan) OperateSSH(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	var req model.SSHOperate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := s.operateSSH(hostID, req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags SSH
// @Summary Generate host SSH secret
// @Description Generate host SSH secret
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param request body model.GenerateKey true "request"
// @Success 200
// @Router /ssh/{host}/keys [post]
func (s *SSHMan) CreateKey(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	var req model.GenerateKey
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := s.createKey(hostID, req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags SSH
// @Summary Get SSH secrets on host
// @Description Get SSH secrets on host
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param keyword query string false "Keyword"
// @Success 200
// @Router /ssh/{host}/keys [get]
func (s *SSHMan) ListKey(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	keyword := c.Query("keyword")

	req := model.ListKey{
		Keyword: keyword,
	}

	data, err := s.listKeys(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags SSH
// @Summary Enable secret on host
// @Description Enable secret on host
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param request body model.EnableKey true "request"
// @Success 200
// @Router /ssh/{host}/keys/enable [post]
func (s *SSHMan) EnableKey(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	var req model.EnableKey
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := s.enableKey(hostID, req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags SSH
// @Summary Enable secret on host
// @Description Enable secret on host
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param request body model.RemoveKey true "request"
// @Success 200
// @Router /ssh/{host}/keys/enable [post]
func (s *SSHMan) RemoveKey(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	var req model.RemoveKey
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := s.removeKey(hostID, req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags SSH
// @Summary Download private key
// @Description Download private key
// @Produce octet-stream
// @Param host path uint true "Host ID"
// @Param source query string true "Source file path"
// @Success 200 {file} binary
// @Router /ssh/{host}/keys/download [get]
func (s *SSHMan) DownloadPrivateKey(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	source := c.Query("source")
	if source == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Source is required", nil)
		return
	}

	if err := s.downloadFile(c, uint(hostID), source); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, err.Error(), err)
		return
	}
}

// @Tags SSH
// @Summary Set secret password
// @Description Set secret password
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param request body model.SetKeyPassword true "request"
// @Success 200
// @Router /ssh/{host}/keys/password/set [post]
func (s *SSHMan) SetKeyPassword(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	var req model.SetKeyPassword
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := s.setKeyPassword(hostID, req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags SSH
// @Summary Update secret password
// @Description Update secret password
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param request body model.UpdateKeyPassword true "request"
// @Success 200
// @Router /ssh/{host}/keys/password/update [post]
func (s *SSHMan) UpdateKeyPassword(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	var req model.UpdateKeyPassword
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := s.updateKeyPassword(hostID, req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags SSH
// @Summary Clear secret password
// @Description Clear secret password
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param request body model.SetKeyPassword true "request"
// @Success 200
// @Router /ssh/{host}/keys/password/clear [post]
func (s *SSHMan) ClearKeyPassword(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	var req model.SetKeyPassword
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := s.clearKeyPassword(hostID, req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags SSH
// @Summary Get SSH logs on host
// @Description Get SSH logs on host
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param page query int true "Page number"
// @Param page_size query int true "Page size"
// @Param info query string false "Info"
// @Param status query string true "Status can be 'Success' or 'Failed' or 'All'"
// @Success 200 {object} model.SSHLog
// @Router /ssh/{host}/logs [get]
func (s *SSHMan) LoadSSHLogs(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	var req model.SearchSSHLog
	if err := helper.CheckQueryAndValidate(&req, c); err != nil {
		return
	}

	data, err := s.loadLog(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithData(c, data)
}

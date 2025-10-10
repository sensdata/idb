package process

import (
	"crypto/tls"
	_ "embed"
	"fmt"
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

type Process struct {
	plugin      plugin.Plugin
	pluginConf  plugin.PluginConf
	restyClient *resty.Client
}

var LOG *log.Log

//go:embed plug.yaml
var plugYAML []byte

//go:embed conf.yaml
var confYAML []byte

func (s *Process) Initialize() {
	global.LOG.Info("process init begin")

	if err := yaml.Unmarshal(plugYAML, &s.plugin); err != nil {
		global.LOG.Error("Failed to load process yaml: %v", err)
		return
	}

	confPath := filepath.Join(constant.CenterConfDir, "process", "conf.yaml")
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
		logger, err := log.InitLogger(s.pluginConf.Items.LogDir, "process.log")
		if err != nil {
			global.LOG.Error("Failed to initialize logger: %v \n", err)
			return
		}
		LOG = logger
	}

	api.API.SetUpPluginRouters(
		"process",
		[]plugin.PluginRoute{
			{Method: "GET", Path: "/:host", Handler: s.GetProcessList},
			{Method: "GET", Path: "/:host/detail", Handler: s.GetProcessDetail},
			{Method: "DELETE", Path: "/:host", Handler: s.DeleteProcess},
		},
	)

	global.LOG.Info("process init end")
}

func (s *Process) Start() {
	global.LOG.Info("process start")
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
}

func (s *Process) Release() {

}

func (s *Process) sendAction(actionRequest model.HostAction) (*model.ActionResponse, error) {
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

// @Tags Process
// @Summary Get process list
// @Description Get process list
// @Accept json
// @Produce json
// @Param host path string true "Host"
// @Param page query int true "Page number"
// @Param page_size query int true "Page size"
// @Param order_by query string true "Order by one of (pid, name, cpu, mem)"
// @Param order query string true "Order, one of (asc, desc)"
// @Param name query string false "Process name"
// @Param pid query int false "Process ID"
// @Param user query string false "User name"
// @Success 200 {object} model.ProcessListResponse
// @Router /process/{host} [get]
func (s *Process) GetProcessList(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	var req model.ProcessListRequest
	if err := helper.CheckQueryAndValidate(&req, c); err != nil {
		return
	}

	result, err := s.getProcessList(uint(hostID), req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}

	helper.SuccessWithData(c, result)
}

// @Tags Process
// @Summary Get process detail
// @Description Get process detail
// @Accept json
// @Produce json
// @Param host path string true "Host"
// @Param pid query int true "Process ID"
// @Success 200 {object} model.ProcessDetail
// @Router /process/{host}/detail [get]
func (s *Process) GetProcessDetail(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	pid, err := strconv.ParseUint(c.Query("pid"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid pid", err)
		return
	}

	var req model.ProcessRequest
	req.PID = int32(pid)

	result, err := s.getProcessDetail(uint(hostID), req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}

	helper.SuccessWithData(c, result)
}

// @Tags Process
// @Summary Delete process
// @Description Delete process
// @Accept json
// @Produce json
// @Param host path string true "Host"
// @Param pid query int true "Process ID"
// @Success 200 {object} model.ActionResponse
// @Router /process/{host} [delete]
func (s *Process) DeleteProcess(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	pid, err := strconv.ParseUint(c.Query("pid"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid pid", err)
		return
	}

	var req model.ProcessRequest
	req.PID = int32(pid)

	err = s.killProcess(uint(hostID), req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}

	helper.SuccessWithData(c, nil)
}

package docker

import (
	"crypto/tls"
	_ "embed"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
	"github.com/sensdata/idb/center/core/api"
	"github.com/sensdata/idb/center/core/api/service"
	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/helper"
	"github.com/sensdata/idb/core/log"
	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/plugin"
	"github.com/sensdata/idb/core/utils"
	"gopkg.in/yaml.v2"
)

type DockerMan struct {
	AppDir string // App目录

	plugin      plugin.Plugin
	pluginConf  plugin.PluginConf
	restyClient *resty.Client
}

var LOG *log.Log

//go:embed plug.yaml
var plugYAML []byte

//go:embed conf.yaml
var confYAML []byte

//go:embed install-docker.sh
var installDockerShell []byte

func (s *DockerMan) Initialize() {
	global.LOG.Info("dockerman init begin \n")

	// 解析plugYAML
	if err := yaml.Unmarshal(plugYAML, &s.plugin); err != nil {
		global.LOG.Error("Failed to load info: %v", err)
		return
	}

	// // 解析formYaml
	// if err := yaml.Unmarshal(formYaml, &s.form); err != nil {
	// 	global.LOG.Error("Failed to load form: %v", err)
	// 	return
	// }

	// // 由templateService解析出模板的templateServiceForm
	// var err error
	// s.templateServiceForm, err = parseServiceBytesToServiceForm(templateService, s.form.Fields)
	// if err != nil {
	// 	global.LOG.Error("Failed to parse template: %v", err)
	// 	return
	// }

	confPath := filepath.Join(constant.CenterConfDir, "docker", "conf.yaml")
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

	global.LOG.Info("dockerman conf: %v", s.pluginConf)

	//初始化日志模块
	if LOG == nil {
		logger, err := log.InitLogger(s.pluginConf.Items.LogDir, "docker.log")
		if err != nil {
			global.LOG.Error("Failed to initialize logger: %v \n", err)
			return
		}
		LOG = logger
	}

	api.API.SetUpPluginRouters(
		"docker",
		[]plugin.PluginRoute{
			{Method: "GET", Path: "/info", Handler: s.GetPluginInfo},
			{Method: "GET", Path: "/menu", Handler: s.GetMenu},

			// docker
			{Method: "GET", Path: "/:host/install/status", Handler: s.DockerInstallStatus}, // 获取docker安装状态
			{Method: "POST", Path: "/:host/install", Handler: s.DockerInstall},             // 安装docker
			{Method: "GET", Path: "/:host/status", Handler: s.DockerStatus},                // 获取docker状态
			{Method: "GET", Path: "/:host/conf", Handler: s.DockerConf},                    // 获取docker配置
			{Method: "PUT", Path: "/:host/conf", Handler: s.DockerUpdateConf},              // 更新docker配置
			{Method: "GET", Path: "/:host/conf/raw", Handler: s.DockerConfRaw},             // 获取docker配置源文
			{Method: "PUT", Path: "/:host/conf/raw", Handler: s.DockerUpdateConfRaw},       // 更新docker配置源文
			{Method: "PUT", Path: "/:host/log", Handler: s.DockerUpdateLogOption},          // 日志设置
			{Method: "PUT", Path: "/:host/ipv6", Handler: s.DockerUpdateIpv6Option},        // ipv6设置
			{Method: "POST", Path: "/:host/operation", Handler: s.DockerOperation},         // 操作docker服务
			{Method: "GET", Path: "/:host/inspect", Handler: s.Inspect},                    // 获取信息（container image volume network）
			{Method: "POST", Path: "/:host/prune", Handler: s.Prune},                       // 清理（container image volume network buildcache）

			// compose
			{Method: "GET", Path: "/:host/compose", Handler: s.ComposeQuery},                // 获取编排列表
			{Method: "GET", Path: "/:host/compose/detail", Handler: s.ComposeDetail},        // 获取编排详情
			{Method: "POST", Path: "/:host/compose", Handler: s.ComposeCreate},              // 创建编排
			{Method: "PUT", Path: "/:host/compose", Handler: s.ComposeUpdate},               // 更新编排
			{Method: "DELETE", Path: "/:host/compose", Handler: s.ComposeDelete},            // 删除编排
			{Method: "POST", Path: "/:host/compose/test", Handler: s.ComposeTest},           // 测试编排
			{Method: "POST", Path: "/:host/compose/operation", Handler: s.ComposeOperation}, // 操作编排
			{Method: "GET", Path: "/:host/compose/logs/tail", Handler: s.FollowComposeLogs}, // 追踪编排日志

			// containers
			{Method: "GET", Path: "/:host/containers", Handler: s.ContainerQuery},                      // 获取容器列表
			{Method: "GET", Path: "/:host/containers/names", Handler: s.ContainerNames},                // 获取容器名列表
			{Method: "GET", Path: "/:host/containers/usages", Handler: s.ContainerUsages},              // 获取容器资源占用
			{Method: "GET", Path: "/:host/containers/usages/follow", Handler: s.ContainerUsagesFollow}, // 追踪容器资源占用
			{Method: "GET", Path: "/:host/containers/usage", Handler: s.ContainerUsage},                // 获取单个容器资源占用
			{Method: "GET", Path: "/:host/containers/usage/follow", Handler: s.ContainerUsageFollow},   // 追踪单个容器资源占用
			{Method: "GET", Path: "/:host/containers/limit", Handler: s.ContainerLimit},                // 获取容器资源限制
			{Method: "POST", Path: "/:host/containers", Handler: s.ContainerCreate},                    // 创建容器
			{Method: "PUT", Path: "/:host/containers", Handler: s.ContainerUpdate},                     // 编辑容器
			{Method: "POST", Path: "/:host/containers/upgrade", Handler: s.ContainerUpgrade},           // 升级容器
			{Method: "POST", Path: "/:host/containers/rename", Handler: s.ContainerRename},             // 重命名容器
			{Method: "POST", Path: "/:host/containers/operation", Handler: s.ContainerOperation},       // 操作容器

			{Method: "GET", Path: "/:host/containers/detail", Handler: s.ContainerInfo},          // 获取容器详情
			{Method: "GET", Path: "/:host/containers/stats", Handler: s.ContainerStats},          // 获取容器监控数据
			{Method: "DELETE", Path: "/:host/containers/logs", Handler: s.ContainerLogClean},     // 清理容器日志
			{Method: "GET", Path: "/:host/containers/logs/tail", Handler: s.FollowContainerLogs}, // 追踪容器日志
			{Method: "GET", Path: "/:host/containers/terminal", Handler: s.HandleTerminal},       // 进入容器终端会话
			{Method: "POST", Path: "/:host/containers/terminal/quit", Handler: s.QuitSession},    // 终止容器终端会话

			// images
			{Method: "GET", Path: "/:host/images", Handler: s.ImagePage},         // 获取镜像列表
			{Method: "GET", Path: "/:host/images/names", Handler: s.ImageNames},  // 获取镜像名列表
			{Method: "POST", Path: "/:host/images/build", Handler: s.ImageBuild}, // 构建镜像
			{Method: "POST", Path: "/:host/images/pull", Handler: s.ImagePull},   // 拉取镜像
			{Method: "POST", Path: "/:host/images/push", Handler: s.ImagePush},   // 推送镜像
			{Method: "POST", Path: "/:host/images/import", Handler: s.ImageLoad}, // 导入镜像
			{Method: "POST", Path: "/:host/images/export", Handler: s.ImageSave}, // 导出镜像
			{Method: "DELETE", Path: "/:host/images", Handler: s.ImageRemove},    // 批量删除镜像
			{Method: "PUT", Path: "/:host/images/tag", Handler: s.ImageTag},      // 设置镜像标签

			// volumes
			{Method: "GET", Path: "/:host/volumes", Handler: s.VolumePage},        // 获取卷列表
			{Method: "GET", Path: "/:host/volumes/names", Handler: s.VolumeNames}, // 获取卷名列表
			{Method: "DELETE", Path: "/:host/volumes", Handler: s.VolumeDelete},   // 批量删除卷
			{Method: "POST", Path: "/:host/volumes", Handler: s.VolumeCreate},     // 创建卷

			// networks
			{Method: "GET", Path: "/:host/networks", Handler: s.NetworkPage},        // 获取网络列表
			{Method: "GET", Path: "/:host/networks/names", Handler: s.NetworkNames}, // 获取网络列表
			{Method: "DELETE", Path: "/:host/networks", Handler: s.NetworkDelete},   // 批量删除网络
			{Method: "POST", Path: "/:host/networks", Handler: s.NetworkCreate},     // 创建网络
		},
	)

	global.LOG.Info("dockerman init end")
}

func (s *DockerMan) Start() {
	global.LOG.Info("dockerman start")
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

func (s *DockerMan) Release() {

}

func (s *DockerMan) sendAction(actionRequest model.HostAction) (*model.ActionResponse, error) {
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

	LOG.Info("action response: %v", actionResponse)

	return &actionResponse, nil
}
func (s *DockerMan) sendCommand(hostId uint, command string) (*model.CommandResult, error) {
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

func (s *DockerMan) createFile(hostID uint64, op model.FileCreate) error {
	data, err := utils.ToJSONString(op)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.File_Create,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		LOG.Error("failed to create file")
		return errors.New("failed to create file")
	}

	return nil
}

// @Tags Docker
// @Summary Get plugin info
// @Description Get plugin information
// @Accept json
// @Produce json
// @Success 200 {object} plugin.PluginInfo
// @Router /docker/info [get]
func (s *DockerMan) GetPluginInfo(c *gin.Context) {
	pluginInfo, err := s.getPluginInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"type": "info", "payload": pluginInfo})
}

// @Tags Docker
// @Summary Get plugin menu
// @Description Get plugin menu items
// @Accept json
// @Produce json
// @Success 200 {array} plugin.MenuItem
// @Router /docker/menu [get]
func (s *DockerMan) GetMenu(c *gin.Context) {
	menuItems, err := s.getMenus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"type": "menu", "payload": menuItems})
}

func (s *DockerMan) getPluginInfo() (plugin.PluginInfo, error) {
	return s.plugin.Info, nil
}

func (s *DockerMan) getMenus() ([]plugin.MenuItem, error) {
	return s.plugin.Menu, nil
}

// @Tags Docker
// @Summary Docker install status
// @Description Get docker install status
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Success 200 {object} model.DockerInstallStatus
// @Router /docker/{host}/install/status [get]
func (s *DockerMan) DockerInstallStatus(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	result, err := s.dockerInstallStatus(hostID)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}

	helper.SuccessWithData(c, result)
}

// @Tags Docker
// @Summary Docker install
// @Description Install docker
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Success 200
// @Router /docker/{host}/install [post]
func (s *DockerMan) DockerInstall(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	err = s.dockerInstall(hostID)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Docker
// @Summary Docker status
// @Description Get docker status
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Success 200 {object} model.DockerStatus
// @Router /docker/{host}/status [get]
func (s *DockerMan) DockerStatus(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	result, err := s.dockerStatus(hostID)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}

	helper.SuccessWithData(c, result)
}

// @Tags Docker
// @Summary Docker configurations
// @Description Get docker configurations
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Success 200 {object} model.DaemonJsonConf
// @Router /docker/{host}/conf [get]
func (s *DockerMan) DockerConf(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	result, err := s.dockerConf(hostID)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}

	helper.SuccessWithData(c, result)
}

// @Tags Docker
// @Summary Update Docker conf
// @Description Update docker conf with key-value
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param request body model.KeyValue true "Configuration key and value"
// @Success 200
// @Router /docker/{host}/conf [put]
func (s *DockerMan) DockerUpdateConf(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	var req model.KeyValue
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	err = s.dockerUpdateConf(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Docker
// @Summary Docker docker configurations raw content
// @Description Get docker configurations raw content
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Success 200 {object} model.DaemonJsonUpdateRaw
// @Router /docker/{host}/conf/raw [get]
func (s *DockerMan) DockerConfRaw(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	result, err := s.dockerConfRaw(hostID)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}

	helper.SuccessWithData(c, result)
}

// @Tags Docker
// @Summary Update Docker conf file
// @Description Update docker conf file with content
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param request body model.DaemonJsonUpdateRaw true "Configuration file details"
// @Success 200
// @Router /docker/{host}/conf/raw [put]
func (s *DockerMan) DockerUpdateConfRaw(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	var req model.DaemonJsonUpdateRaw
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	err = s.dockerUpdateConfByFile(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Docker
// @Summary Update Docker log option
// @Description Update docker log option
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param request body model.LogOption true "Configuration key and value"
// @Success 200
// @Router /docker/{host}/log [put]
func (s *DockerMan) DockerUpdateLogOption(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	var req model.LogOption
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	err = s.dockerUpdateLogOption(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Docker
// @Summary Update Docker ipv6 option
// @Description Update docker ipv6 option
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param request body model.Ipv6Option true "Configuration key and value"
// @Success 200
// @Router /docker/{host}/ipv6 [put]
func (s *DockerMan) DockerUpdateIpv6Option(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	var req model.Ipv6Option
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	err = s.dockerUpdateIpv6Option(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Docker
// @Summary Docker operations
// @Description To start, stop, restart docker
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param request body model.DockerOperation true "Operation options"
// @Success 200
// @Router /docker/{host}/operation [post]
func (s *DockerMan) DockerOperation(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	var req model.DockerOperation
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	err = s.dockerOperation(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Docker
// @Summary Inspect
// @Description Get details of a container, image, volume or network
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param type query string true "Type of inspection, can be one of container, image, volume and network"
// @Param id query string true "ID of the object"
// @Success 200
// @Router /docker/{host}/inspect [get]
func (s *DockerMan) Inspect(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	objectType := c.Query("type")
	if objectType == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid type", err)
		return
	}

	objectID := c.Query("id")
	if objectID == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid id", err)
		return
	}

	req := model.Inspect{
		Type: objectType,
		ID:   objectID,
	}

	result, err := s.inspect(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}

	helper.SuccessWithData(c, result)
}

// @Tags Docker
// @Summary Prune
// @Description Prune operation for container, image, volume, network or buildcache
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param request body model.Prune true "Prune details"
// @Success 200 {object} model.PruneResult
// @Router /docker/{host}/prune [post]
func (s *DockerMan) Prune(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	var req model.Prune
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	result, err := s.prune(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}

	helper.SuccessWithData(c, result)
}

// @Tags Docker
// @Summary Query compose
// @Description Query compose
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param info query string false "Info for searching"
// @Param page query int true "Page number"
// @Param page_size query int true "Page size"
// @Success 200 {object} model.PageResult
// @Router /docker/{host}/compose [get]
func (s *DockerMan) ComposeQuery(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	var req model.QueryCompose
	if err := helper.CheckQueryAndValidate(&req, c); err != nil {
		return
	}
	req.WorkDir = s.AppDir
	result, err := s.composeQuery(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}

	helper.SuccessWithData(c, result)
}

// @Tags Docker
// @Summary Create compose
// @Description Create compose
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param request body model.CreateCompose true "Compose creation details"
// @Success 200 {object} model.ComposeCreateResult
// @Router /docker/{host}/compose [post]
func (s *DockerMan) ComposeCreate(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	var req model.CreateCompose
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	result, err := s.composeCreate(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}

	helper.SuccessWithData(c, result)
}

// @Tags Docker
// @Summary Get compose detail
// @Description Get compose detail
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param name query string true "Compose name"
// @Success 200 {object} model.ComposeDetailRsp
// @Router /docker/{host}/compose/detail [get]
func (s *DockerMan) ComposeDetail(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	name := c.Query("name")
	if name == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid name", err)
		return
	}

	req := model.ComposeDetailReq{
		Name:    name,
		WorkDir: s.AppDir,
	}
	result, err := s.composeDetail(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	helper.SuccessWithData(c, result)
}

// @Tags Docker
// @Summary Update compose
// @Description Update compose
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param request body model.CreateCompose true "Compose edit details"
// @Success 200 {object} model.ComposeCreateResult
// @Router /docker/{host}/compose [put]
func (s *DockerMan) ComposeUpdate(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	var req model.CreateCompose
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	result, err := s.composeUpdate(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}

	helper.SuccessWithData(c, result)
}

// @Tags Docker
// @Summary Remove compose
// @Description Remove compose
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param name query string true "Compose name"
// @Success 200 {object} model.ComposeCreateResult
// @Router /docker/{host}/compose [delete]
func (s *DockerMan) ComposeDelete(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	name := c.Query("name")
	if name == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid name", err)
		return
	}

	req := model.ComposeRemove{Name: name, WorkDir: s.AppDir}
	result, err := s.composeRemove(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}

	helper.SuccessWithData(c, result)
}

// @Tags Docker
// @Summary Test compose
// @Description Test compose
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param request body model.CreateCompose true "Compose creation details"
// @Success 200 {object} model.ComposeTestResult
// @Router /docker/{host}/compose/test [post]
func (s *DockerMan) ComposeTest(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	var req model.CreateCompose
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	result, err := s.composeTest(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}

	helper.SuccessWithData(c, result)
}

// @Tags Docker
// @Summary Operate compose
// @Description Operate compose
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param request body model.OperateCompose true "Compose operation details"
// @Success 200 {object} model.OperationResult
// @Router /docker/{host}/compose/operation [post]
func (s *DockerMan) ComposeOperation(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	var req model.OperateCompose
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	result, err := s.composeOperation(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}

	helper.SuccessWithData(c, result)
}

// @Tags Docker
// @Summary Connect to compose log stream
// @Description Connect to a compose log stream through SSE
// @Accept json
// @Produce text/event-stream
// @Param host path int true "Host ID"
// @Param config_files query string true "Comma-separated list of config files to tail"
// @Success 200 {string} string "SSE stream started"
// @Failure 400 {object} model.Response "Bad Request"
// @Router /docker/{host}/compose/logs/tail [get]
func (s *DockerMan) FollowComposeLogs(c *gin.Context) {
	err := s.followComposeLogs(c)
	if err != nil {
		global.LOG.Error("Handle compose log stream failed: %v", err)
		helper.ErrorWithDetail(c, http.StatusInternalServerError, "Failed to establish SSE connection", err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Docker
// @Summary Query containers
// @Description Query containers
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param info query string false "Info for searching"
// @Param state query string true "Container state, one of (all created running paused restarting removing exited dead)"
// @Param page query int true "Page number"
// @Param page_size query int true "Page size"
// @Param order_by query string false "Order by one of (name, state, created)"
// @Success 200 {object} model.PageResult
// @Router /docker/{host}/containers [get]
func (s *DockerMan) ContainerQuery(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	var req model.QueryContainer
	if err := helper.CheckQueryAndValidate(&req, c); err != nil {
		return
	}

	result, err := s.containerQuery(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}

	helper.SuccessWithData(c, result)
}

// @Tags Docker
// @Summary Query container names
// @Description Query container names
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Success 200 {object} model.PageResult
// @Router /docker/{host}/containers/names [get]
func (s *DockerMan) ContainerNames(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	result, err := s.containerNames(hostID)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}

	helper.SuccessWithData(c, result)
}

// @Tags Docker
// @Summary Get container resource usage list
// @Description Get container resource usage list
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param info query string false "Info for searching"
// @Param state query string true "Container state, one of (all created running paused restarting removing exited dead)"
// @Param page query int true "Page number"
// @Param page_size query int true "Page size"
// @Param order_by query string false "Order by one of (name, state, created)"
// @Success 200 {object} model.PageResult
// @Router /docker/{host}/containers/usages [get]
func (s *DockerMan) ContainerUsages(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	var req model.QueryContainer
	if err := helper.CheckQueryAndValidate(&req, c); err != nil {
		return
	}

	result, err := s.containerUsages(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}

	helper.SuccessWithData(c, result)
}

// @Tags Docker
// @Summary Follow container resource usage list
// @Description Follow container resource usage list
// @Accept json
// @Produce text/event-stream
// @Param host path int true "Host ID"
// @Param info query string false "Info for searching"
// @Param state query string true "Container state, one of (all created running paused restarting removing exited dead)"
// @Param page query int true "Page number"
// @Param page_size query int true "Page size"
// @Param order_by query string false "Order by one of (name, state, created)"
// @Success 200 {string} string "SSE stream started"
// @Failure 400 {object} model.Response "Bad Request"
// @Router /docker/{host}/containers/usages/follow [get]
func (s *DockerMan) ContainerUsagesFollow(c *gin.Context) {
	err := s.followContainerUsages(c)
	if err != nil {
		global.LOG.Error("Handle container resouce usage list stream failed: %v", err)
		helper.ErrorWithDetail(c, http.StatusInternalServerError, "Failed to establish SSE connection", err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Docker
// @Summary Get resource usage of the specified container
// @Description Get resource usage of the specified container
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param id query string true "Container ID"
// @Success 200 {object} model.PageResult
// @Router /docker/{host}/containers/usage [get]
func (s *DockerMan) ContainerUsage(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	containerID := c.Query("id")
	if containerID == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid container id", err)
		return
	}

	result, err := s.containerUsage(hostID, containerID)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}

	helper.SuccessWithData(c, result)
}

// @Tags Docker
// @Summary Follow resource usage of the specified container
// @Description Follow resource usage of the specified container
// @Accept json
// @Produce text/event-stream
// @Param host path int true "Host ID"
// @Param id query string true "Container ID"
// @Success 200 {string} string "SSE stream started"
// @Failure 400 {object} model.Response "Bad Request"
// @Router /docker/{host}/containers/usage/follow [get]
func (s *DockerMan) ContainerUsageFollow(c *gin.Context) {
	err := s.followContainerUsage(c)
	if err != nil {
		global.LOG.Error("Handle container resouce usage stream failed: %v", err)
		helper.ErrorWithDetail(c, http.StatusInternalServerError, "Failed to establish SSE connection", err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Docker
// @Summary Get container resource usage list
// @Description Get container resource usage list
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Success 200 {object} model.ContainerResourceLimit
// @Router /docker/{host}/containers/limit [get]
func (s *DockerMan) ContainerLimit(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	result, err := s.containerLimit(hostID)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}

	helper.SuccessWithData(c, result)
}

// @Tags Docker
// @Summary Create container
// @Description 创建容器
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param request body model.ContainerOperate true "Container creation details"
// @Success 200
// @Router /docker/{host}/containers [post]
func (s *DockerMan) ContainerCreate(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	var req model.ContainerOperate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	err = s.createContainer(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Docker
// @Summary Update Container
// @Description Update detail of the specified container
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param request body model.ContainerOperate true "Container edit details"
// @Success 200
// @Router /docker/{host}/containers [put]
func (s *DockerMan) ContainerUpdate(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	var req model.ContainerOperate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	err = s.updateContainer(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Docker
// @Summary Upgrade container
// @Description Upgrade container
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param request body model.ContainerUpgrade true "Container upgrade details"
// @Success 200
// @Router /docker/{host}/containers/upgrade [post]
func (s *DockerMan) ContainerUpgrade(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	var req model.ContainerUpgrade
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	err = s.upgradeContainer(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Docker
// @Summary Rename container
// @Description Rename container
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param request body model.Rename true "Container rename details"
// @Success 200
// @Router /docker/{host}/containers/rename [post]
func (s *DockerMan) ContainerRename(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	var req model.Rename
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	err = s.renameContainer(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Docker
// @Summary Execute operations to container
// @Description Execute operations to container
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param request body model.ContainerOperation true "Container operation details"
// @Success 200 {object} model.OperationResult
// @Router /docker/{host}/containers/operation [post]
func (s *DockerMan) ContainerOperation(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	var req model.ContainerOperation
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	result, err := s.operateContainer(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}

	helper.SuccessWithData(c, result)
}

// @Tags Docker
// @Summary Get Container info
// @Description Get detail of the specified container
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param id query string true "Container ID"
// @Success 200 {object} model.ContainerOperate
// @Router /docker/{host}/containers/detail [get]
func (s *DockerMan) ContainerInfo(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	containerID := c.Query("id")
	if containerID == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid container id", err)
		return
	}

	result, err := s.containerInfo(hostID, containerID)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}

	helper.SuccessWithData(c, result)
}

// @Tags Docker
// @Summary Get Container stats
// @Description Get stats of the specified container
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param id query string true "Container ID"
// @Success 200 {object} model.ContainerStats
// @Router /docker/{host}/containers/stats [get]
func (s *DockerMan) ContainerStats(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	containerID := c.Query("id")
	if containerID == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid container id", err)
		return
	}

	result, err := s.containerStats(hostID, containerID)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}

	helper.SuccessWithData(c, result)
}

// @Tags Docker
// @Summary Clean Container logs
// @Description Clean container logs
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param id query string true "Container ID"
// @Success 200
// @Router /docker/{host}/containers/logs [delete]
func (s *DockerMan) ContainerLogClean(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	containerID := c.Query("id")
	if containerID == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid container id", err)
		return
	}

	err = s.containerLogClean(hostID, containerID)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Docker
// @Summary Connect to  container log stream
// @Description Connect to a container log stream through SSE
// @Accept json
// @Produce text/event-stream
// @Param host path int true "Host ID"
// @Param id query string true "Container ID"
// @Param follow query bool false "Follow the log stream"
// @Param tail query int false "How many lines from the end of the logs to show, can be one of 0, 100, 200, 500, 1000. Default is 100. Pass 0 to show all logs."
// @Param since query string false "Show logs since a certain time, can be one of 24h, 4h, 1h, 10m. If not specified, all logs will be shown."
// @Success 200 {string} string "SSE stream started"
// @Failure 400 {object} model.Response "Bad Request"
// @Router /docker/{host}/containers/logs/tail [get]
func (s *DockerMan) FollowContainerLogs(c *gin.Context) {
	err := s.followContainerLogs(c)
	if err != nil {
		global.LOG.Error("Handle container log stream failed: %v", err)
		helper.ErrorWithDetail(c, http.StatusInternalServerError, "Failed to establish SSE connection", err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Docker
// @Summary Create or reconnect to a container terminal session through websocket
// @Description Create or reconnect to a container terminal session through websocket.
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param cols query uint false "Window cols, default 80"
// @Param rows query uint false "Window rows, default 40"
// @Success 200
// @Router /docker/{host}/containers/terminal [get]
func (s *DockerMan) HandleTerminal(c *gin.Context) {
	err := s.handleContainerTerminal(c)
	if err != nil {
		global.LOG.Error("Handle container terminal failed: %v", err)

		// 检查是否为升级错误
		if err.Error() == "websocket: the client is not using the websocket protocol: 'upgrade' token not found in 'Connection' header" {
			helper.ErrorWithDetail(c, http.StatusBadRequest, "Failed to establish WebSocket connection", err)
		} else {
			// 对于其他错误，返回一个不同的状态码
			helper.ErrorWithDetail(c, http.StatusInternalServerError, "Internal server error", err)
		}
		return
	}
}

// @Tags Docker
// @Summary Quit container terminal session
// @Description Quit container terminal session
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param request body model.TerminalRequest true "Request details"
// @Success 200
// @Router /docker/{host}/containers/terminal/quit [post]
func (s *DockerMan) QuitSession(c *gin.Context) {
	token := c.GetHeader("Authorization")

	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	var req model.TerminalRequest
	if err := helper.CheckBind(&req, c); err != nil {
		return
	}

	err = s.quitContainerSession(token, uint(hostID), req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}
	helper.SuccessWithData(c, "")
}

// @Tags Docker
// @Summary Get images
// @Description Get images
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param info query string false "Info for searching"
// @Param page query int true "Page number"
// @Param page_size query int true "Page size"
// @Success 200 {object} model.PageResult
// @Router /docker/{host}/images [get]
func (s *DockerMan) ImagePage(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	var req model.SearchPageInfo
	if err := helper.CheckQueryAndValidate(&req, c); err != nil {
		return
	}

	result, err := s.getImages(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}

	helper.SuccessWithData(c, result)
}

// @Tags Docker
// @Summary Query image names
// @Description Query image names
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Success 200 {object} model.PageResult
// @Router /docker/{host}/images/names [get]
func (s *DockerMan) ImageNames(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	result, err := s.imageNames(hostID)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}

	helper.SuccessWithData(c, result)
}

// @Tags Docker
// @Summary Build image
// @Description Build image
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param request body model.ImageBuild true "Build image details"
// @Success 200 {object} model.OperationResult
// @Router /docker/{host}/images/build [post]
func (s *DockerMan) ImageBuild(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	var req model.ImageBuild
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	result, err := s.buildImage(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}

	helper.SuccessWithData(c, result)
}

// @Tags Docker
// @Summary Pull image
// @Description Pull image
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param request body model.ImagePull true "Pull image details"
// @Success 200 {object} model.OperationResult
// @Router /docker/{host}/images/pull [post]
func (s *DockerMan) ImagePull(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	var req model.ImagePull
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	result, err := s.pullImage(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}

	helper.SuccessWithData(c, result)
}

// @Tags Docker
// @Summary Import image
// @Description Import image
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param request body model.ImageLoad true "Import image details"
// @Success 200 {object} model.OperationResult
// @Router /docker/{host}/images/import [post]
func (s *DockerMan) ImageLoad(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	var req model.ImageLoad
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	result, err := s.loadImage(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}

	helper.SuccessWithData(c, result)
}

// @Tags Docker
// @Summary Export image
// @Description Export image
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param request body model.ImageSave true "Export image details"
// @Success 200 {object} model.OperationResult
// @Router /docker/{host}/images/export [post]
func (s *DockerMan) ImageSave(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	var req model.ImageSave
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	result, err := s.exportImage(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}

	helper.SuccessWithData(c, result)
}

// @Tags Docker
// @Summary Push image
// @Description Push image
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param id path string true "Image ID"
// @Param request body model.ImagePush true "Push image details"
// @Success 200 {object} model.OperationResult
// @Router /docker/{host}/images/push [post]
func (s *DockerMan) ImagePush(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	var req model.ImagePush
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	result, err := s.pushImage(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}

	helper.SuccessWithData(c, result)
}

// @Tags Docker
// @Summary Batch Delete images
// @Description Batch Delete images
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param force query bool true "Force delete"
// @Param sources query string true "Comma-separated list of image names to delete"
// @Success 200 {object} model.OperationResult
// @Router /docker/{host}/images [delete]
func (s *DockerMan) ImageRemove(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	force, _ := strconv.ParseBool(c.Query("force"))

	sources := c.Query("sources")
	if sources == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "No sources provided", nil)
		return
	}

	req := model.BatchDelete{
		Force: force,
		Names: strings.Split(sources, ","),
	}

	result, err := s.deleteImage(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}

	helper.SuccessWithData(c, result)
}

// @Tags Docker
// @Summary Set image tag
// @Description Set image tag
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param request body model.ImageTag true "Set image tag details"
// @Success 200
// @Router /docker/{host}/images/tag [put]
func (s *DockerMan) ImageTag(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	var req model.ImageTag
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	err = s.setImageTag(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Docker
// @Summary Get volumes
// @Description Get volumes
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param info query string false "Info for searching"
// @Param page query int true "Page number"
// @Param page_size query int true "Page size"
// @Success 200 {object} model.PageResult
// @Router /docker/{host}/volumes [get]
func (s *DockerMan) VolumePage(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	var req model.SearchPageInfo
	if err := helper.CheckQueryAndValidate(&req, c); err != nil {
		return
	}

	result, err := s.getVolumes(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}

	helper.SuccessWithData(c, result)
}

// @Tags Docker
// @Summary Query volume names
// @Description Query volume names
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Success 200 {object} model.PageResult
// @Router /docker/{host}/volumes/names [get]
func (s *DockerMan) VolumeNames(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	result, err := s.volumeNames(hostID)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}

	helper.SuccessWithData(c, result)
}

// @Tags Docker
// @Summary Batch Delete volume
// @Description Batch Delete volume
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param force query bool true "Force delete"
// @Param sources query string true "Comma-separated list of volume ids to delete"
// @Success 200
// @Router /docker/{host}/volumes [delete]
func (s *DockerMan) VolumeDelete(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	force, _ := strconv.ParseBool(c.Query("force"))

	sources := c.Query("sources")
	if sources == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "No sources provided", nil)
		return
	}

	req := model.BatchDelete{
		Force: force,
		Names: strings.Split(sources, ","),
	}

	err = s.deleteVolume(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Docker
// @Summary Create volume
// @Description Create volume
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param request body model.VolumeCreate true "Create volume details"
// @Success 200
// @Router /docker/{host}/volumes [post]
func (s *DockerMan) VolumeCreate(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	var req model.VolumeCreate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	err = s.createVolume(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Docker
// @Summary Get networks
// @Description Get networks
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param info query string false "Info for searching"
// @Param page query int true "Page number"
// @Param page_size query int true "Page size"
// @Success 200 {object} model.PageResult
// @Router /docker/{host}/networks [get]
func (s *DockerMan) NetworkPage(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	var req model.SearchPageInfo
	if err := helper.CheckQueryAndValidate(&req, c); err != nil {
		return
	}

	result, err := s.getNetworks(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}

	helper.SuccessWithData(c, result)
}

// @Tags Docker
// @Summary Query network names
// @Description Query network names
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Success 200 {object} model.PageResult
// @Router /docker/{host}/networks/names [get]
func (s *DockerMan) NetworkNames(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	result, err := s.networkNames(hostID)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}

	helper.SuccessWithData(c, result)
}

// @Tags Docker
// @Summary Batch Delete network
// @Description Batch Delete network
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param force query bool true "Force delete"
// @Param sources query string true "Comma-separated list of volume ids to delete"
// @Success 200
// @Router /docker/{host}/networks [delete]
func (s *DockerMan) NetworkDelete(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	force, _ := strconv.ParseBool(c.Query("force"))

	sources := c.Query("sources")
	if sources == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "No sources provided", nil)
		return
	}

	req := model.BatchDelete{
		Force: force,
		Names: strings.Split(sources, ","),
	}

	err = s.deleteNetwork(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Docker
// @Summary Create network
// @Description Create network
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param request body model.NetworkCreate true "Create network details"
// @Success 200
// @Router /docker/{host}/networks [post]
func (s *DockerMan) NetworkCreate(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	var req model.NetworkCreate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	err = s.createNetwork(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}

	helper.SuccessWithData(c, nil)
}

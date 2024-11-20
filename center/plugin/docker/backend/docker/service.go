package docker

import (
	_ "embed"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/sensdata/idb/center/core/api"
	"github.com/sensdata/idb/center/core/conn"
	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/helper"
	"github.com/sensdata/idb/core/log"
	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/plugin"
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

	baseUrl := fmt.Sprintf("http://%s:%d/api/v1", "127.0.0.1", conn.CONFMAN.GetConfig().Port)

	s.restyClient = resty.New().
		SetBaseURL(baseUrl).
		SetHeader("Content-Type", "application/json")

	api.API.SetUpPluginRouters(
		"docker",
		[]plugin.PluginRoute{
			{Method: "GET", Path: "/info", Handler: s.GetPluginInfo},
			{Method: "GET", Path: "/menu", Handler: s.GetMenu},
			// {Method: "GET", Path: "/apps", Handler: s.GetApps}, // 获取应用列表

			// containers
			{Method: "GET", Path: "/containers/:host", Handler: s.GetContainers},                      // 获取容器列表
			{Method: "POST", Path: "/containers/:host", Handler: s.CreateContainer},                   // 创建容器
			{Method: "POST", Path: "/containers/:host/prune", Handler: s.PruneContainer},              // 清理容器
			{Method: "GET", Path: "/containers/:host/:id", Handler: s.GetContainer},                   // 获取容器详情
			{Method: "PUT", Path: "/containers/:host/:id", Handler: s.UpdateContainer},                // 编辑容器
			{Method: "DELETE", Path: "/containers/:host/:id", Handler: s.DeleteContainer},             // 删除容器
			{Method: "GET", Path: "/containers/:host/:id/log", Handler: s.GetContainerLog},            // 获取容器日志
			{Method: "GET", Path: "/containers/:host/:id/status", Handler: s.GetContainerStatus},      // 获取容器监控数据
			{Method: "POST", Path: "/containers/:host/:id/upgrade", Handler: s.UpgradeContainer},      // 升级容器
			{Method: "POST", Path: "/containers/:host/:id/start", Handler: s.StartContainer},          // 启动容器
			{Method: "POST", Path: "/containers/:host/:id/stop", Handler: s.StopContainer},            // 停止容器
			{Method: "POST", Path: "/containers/:host/:id/stop/force", Handler: s.ForceStopContainer}, // 强制停止容器
			{Method: "POST", Path: "/containers/:host/:id/reboot", Handler: s.RebootContainer},        // 暂停容器
			{Method: "POST", Path: "/containers/:host/:id/pause", Handler: s.PauseContainer},          // 暂停容器
			{Method: "POST", Path: "/containers/:host/:id/resume", Handler: s.ResumeContainer},        // 恢复容器

			// images
			{Method: "GET", Path: "/images/:host", Handler: s.GetImages},               // 获取镜像列表
			{Method: "POST", Path: "/images/:host/prune", Handler: s.PruneImage},       // 清理镜像
			{Method: "POST", Path: "/images/:host/pull", Handler: s.PullImage},         // 拉取镜像
			{Method: "POST", Path: "/images/:host/import", Handler: s.ImportImage},     // 导入镜像
			{Method: "POST", Path: "/images/:host/build", Handler: s.BuildImage},       // 构建镜像
			{Method: "POST", Path: "/images/:host/build/clean", Handler: s.CleanBuild}, // 清理构建缓存
			{Method: "GET", Path: "/images/:host/:id", Handler: s.GetImage},            // 获取镜像详情
			{Method: "POST", Path: "/images/:host/:id/push", Handler: s.PushImage},     // 推送镜像
			{Method: "POST", Path: "/images/:host/:id/export", Handler: s.ExportImage}, // 导出镜像
			{Method: "PUT", Path: "/images/:host/:id/tag", Handler: s.SetImageTag},     // 设置镜像标签
			{Method: "DELETE", Path: "/images/:host/:id", Handler: s.DeleteImage},      // 删除镜像

			// volumes
			{Method: "GET", Path: "/volumes/:host", Handler: s.GetVolumes},          // 获取卷列表
			{Method: "POST", Path: "/volumes/:host", Handler: s.CreateVolume},       // 创建卷
			{Method: "POST", Path: "/volumes/:host/prune", Handler: s.PruneVolume},  // 清理卷
			{Method: "GET", Path: "/volumes/:host/:id", Handler: s.GetVolume},       // 获取卷详情
			{Method: "DELETE", Path: "/volumes/:host/:id", Handler: s.DeleteVolume}, // 删除卷

			// networks
			{Method: "GET", Path: "/networks/:host", Handler: s.GetNetworks},          // 获取网络列表
			{Method: "POST", Path: "/networks/:host", Handler: s.CreateNetwork},       // 创建网络
			{Method: "POST", Path: "/networks/:host/prune", Handler: s.PruneNetwork},  // 清理网络
			{Method: "GET", Path: "/networks/:host/:id", Handler: s.GetNetwork},       // 获取网络详情
			{Method: "DELETE", Path: "/networks/:host/:id", Handler: s.DeleteNetwork}, // 删除网络
		},
	)

	global.LOG.Info("dockerman init end")
}

func (s *DockerMan) Release() {

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
// @Summary List containers
// @Description 获取容器列表
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param page query int true "Page number"
// @Param page_size query int true "Page size"
// @Success 200 {object} model.PageResult
// @Router /containers/:host [get]
func (s *DockerMan) GetContainers(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	result, err := s.getContainers(hostID)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
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
// @Param request body model.CreateContainer true "Container creation details"
// @Success 200
// @Router /containers/:host [post]
func (s *DockerMan) CreateContainer(c *gin.Context) {
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
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Docker
// @Summary Prune container
// @Description Prune container
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Success 200
// @Router /containers/:host/prune [post]
func (s *DockerMan) PruneContainer(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	err = s.pruneContainer(hostID)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Docker
// @Summary Get Container
// @Description Get detail of the specified container
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param id path int true "Container ID"
// @Success 200 {object} model.Container
// @Router /containers/:host/:id [get]
func (s *DockerMan) GetContainer(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	containerID := c.Param("id")
	if containerID == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid container id", err)
		return
	}

	result, err := s.getContainer(hostID, containerID)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}

	helper.SuccessWithData(c, result)
}

// @Tags Docker
// @Summary Update Container
// @Description Update detail of the specified container
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param id path int true "Container ID"
// @Param request body model.UpdateContainer true "Container edit details"
// @Success 200 {object} model.Container
// @Router /containers/:host/:id [put]
func (s *DockerMan) UpdateContainer(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	containerID := c.Param("id")
	if containerID == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid container id", err)
		return
	}

	var req model.ContainerOperate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	err = s.updateContainer(hostID, containerID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Docker
// @Summary Delete Container
// @Description Delete the specified container
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param id path int true "Container ID"
// @Success 200
// @Router /containers/:host/:id [delete]
func (s *DockerMan) DeleteContainer(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	containerID := c.Param("id")
	if containerID == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid container id", err)
		return
	}

	err = s.deleteContainer(hostID, containerID)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Docker
// @Summary Get container log
// @Description Get container log
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param id path int true "Container ID"
// @Success 200 {object} model.ContainerLog
// @Router /containers/:host/:id/log [get]
func (s *DockerMan) GetContainerLog(c *gin.Context) {
	// hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	// if err != nil {
	// 	helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
	// 	return
	// }

	// containerID := c.Param("id")
	// if containerID == "" {
	// 	helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid container id", err)
	// 	return
	// }

	// result, err := s.getContainerLog(hostID, containerID)
	// if err != nil {
	// 	helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
	// 	return
	// }

	helper.SuccessWithData(c, nil)
}

// @Tags Docker
// @Summary Get container status
// @Description Get container status
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param id path int true "Container ID"
// @Success 200 {object} model.ContainerStatus
// @Router /containers/:host/:id/status [get]
func (s *DockerMan) GetContainerStatus(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	containerID := c.Param("id")
	if containerID == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid container id", err)
		return
	}

	result, err := s.getContainerStatus(hostID, containerID)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}

	helper.SuccessWithData(c, result)
}

// @Tags Docker
// @Summary Upgrade container
// @Description Upgrade container
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param id path int true "Container ID"
// @Success 200
// @Router /containers/:host/:id/upgrade [post]
func (s *DockerMan) UpgradeContainer(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	containerID := c.Param("id")
	if containerID == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid container id", err)
		return
	}

	err = s.upgradeContainer(hostID, containerID)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Docker
// @Summary Start container
// @Description Start container
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param id path int true "Container ID"
// @Success 200
// @Router /containers/:host/:id/start [post]
func (s *DockerMan) StartContainer(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	containerID := c.Param("id")
	if containerID == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid container id", err)
		return
	}

	err = s.startContainer(hostID, containerID)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Docker
// @Summary Stop container
// @Description Stop container
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param id path int true "Container ID"
// @Success 200
// @Router /containers/:host/:id/stop [post]
func (s *DockerMan) StopContainer(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	containerID := c.Param("id")
	if containerID == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid container id", err)
		return
	}

	err = s.stopContainer(hostID, containerID, false)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Docker
// @Summary Force Stop container
// @Description Force Stop container
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param id path int true "Container ID"
// @Success 200
// @Router /containers/:host/:id/stop/force [post]
func (s *DockerMan) ForceStopContainer(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	containerID := c.Param("id")
	if containerID == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid container id", err)
		return
	}

	err = s.stopContainer(hostID, containerID, true)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Docker
// @Summary Reboot container
// @Description Reboot container
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param id path int true "Container ID"
// @Success 200
// @Router /containers/:host/:id/reboot [post]
func (s *DockerMan) RebootContainer(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	containerID := c.Param("id")
	if containerID == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid container id", err)
		return
	}

	err = s.rebootContainer(hostID, containerID)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Docker
// @Summary Pause container
// @Description Pause container
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param id path int true "Container ID"
// @Success 200
// @Router /containers/:host/:id/pause [post]
func (s *DockerMan) PauseContainer(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	containerID := c.Param("id")
	if containerID == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid container id", err)
		return
	}

	err = s.pauseContainer(hostID, containerID)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Docker
// @Summary Resume container
// @Description Resume container
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param id path int true "Container ID"
// @Success 200
// @Router /containers/:host/:id/resume [post]
func (s *DockerMan) ResumeContainer(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	containerID := c.Param("id")
	if containerID == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid container id", err)
		return
	}

	err = s.resumeContainer(hostID, containerID)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Docker
// @Summary Get images
// @Description Get images
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param page query int true "Page number"
// @Param page_size query int true "Page size"
// @Success 200 {object} model.PageResult
// @Router /images/:host [get]
func (s *DockerMan) GetImages(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	result, err := s.getImages(hostID)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}

	helper.SuccessWithData(c, result)
}

// @Tags Docker
// @Summary Prune images
// @Description Prune images
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param request body model.PruneImage true "Prune image details"
// @Success 200
// @Router /images/:host/prune [post]
func (s *DockerMan) PruneImage(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	var req model.Prune
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	err = s.pruneImage(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Docker
// @Summary Pull image
// @Description Pull image
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param request body model.PullImage true "Pull image details"
// @Success 200
// @Router /images/:host/pull [post]
func (s *DockerMan) PullImage(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	var req model.ImagePull
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	err = s.pullImage(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Docker
// @Summary Import image
// @Description Import image
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param request body model.ImportImage true "Import image details"
// @Success 200
// @Router /images/:host/import [post]
func (s *DockerMan) ImportImage(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	var req model.ImageLoad
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	err = s.importImage(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Docker
// @Summary Build image
// @Description Build image
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param request body model.BuildImage true "Build image details"
// @Success 200
// @Router /images/:host/build [post]
func (s *DockerMan) BuildImage(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	var req model.ImageBuild
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	err = s.buildImage(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Docker
// @Summary Clean Build Cache
// @Description Clean build cache
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Success 200
// @Router /images/:host/build/clean [post]
func (s *DockerMan) CleanBuild(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	err = s.cleanBuild(hostID)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Docker
// @Summary Get image
// @Description Get image
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param id path string true "Image ID"
// @Param page query int true "Page number"
// @Param page_size query int true "Page size"
// @Success 200 {object} model.PageResult
// @Router /images/:host/:id [get]
func (s *DockerMan) GetImage(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	imageID := c.Param("id")
	if imageID == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid image id", err)
		return
	}

	result, err := s.getImage(hostID, imageID)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
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
// @Param request body model.PushImage true "Push image details"
// @Success 200
// @Router /images/:host/:id/push [post]
func (s *DockerMan) PushImage(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	imageID := c.Param("id")
	if imageID == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid image id", err)
		return
	}

	var req model.ImagePush
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	err = s.pushImage(hostID, imageID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Docker
// @Summary Export image
// @Description Export image
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param id path string true "Image ID"
// @Param request body model.ExportImage true "Push image details"
// @Success 200
// @Router /images/:host/:id/export [post]
func (s *DockerMan) ExportImage(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	imageID := c.Param("id")
	if imageID == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid image id", err)
		return
	}

	var req model.ImageSave
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	err = s.exportImage(hostID, imageID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Docker
// @Summary Set image tag
// @Description Set image tag
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param id path string true "Image ID"
// @Param request body model.SetImageTag true "Push image details"
// @Success 200
// @Router /images/:host/:id/tag [put]
func (s *DockerMan) SetImageTag(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	imageID := c.Param("id")
	if imageID == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid image id", err)
		return
	}

	var req model.ImageTag
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	err = s.setImageTag(hostID, imageID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Docker
// @Summary Delete image
// @Description Delete image
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param id path string true "Image ID"
// @Success 200
// @Router /images/:host/:id [delete]
func (s *DockerMan) DeleteImage(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	imageID := c.Param("id")
	if imageID == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid image id", err)
		return
	}

	err = s.deleteImage(hostID, imageID)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
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
// @Param page query int true "Page number"
// @Param page_size query int true "Page size"
// @Success 200 {object} model.PageResult
// @Router /volumes/:host [get]
func (s *DockerMan) GetVolumes(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	result, err := s.getVolumes(hostID)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}

	helper.SuccessWithData(c, result)
}

// @Tags Docker
// @Summary Create volume
// @Description Create volume
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param request body model.CreateVolume true "Create volume details"
// @Success 200
// @Router /volumes/:host [post]
func (s *DockerMan) CreateVolume(c *gin.Context) {
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
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Docker
// @Summary Prune volume
// @Description Prune volume
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Success 200
// @Router /volumes/:host/prune [post]
func (s *DockerMan) PruneVolume(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	err = s.pruneVolume(hostID)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Docker
// @Summary Get volume
// @Description Get volume
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param id path string true "Volume ID"
// @Success 200 {object} model.Volume
// @Router /volumes/:host/:id [get]
func (s *DockerMan) GetVolume(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	volumeID := c.Param("id")
	if volumeID == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid volume id", err)
		return
	}

	result, err := s.getVolume(hostID, volumeID)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}

	helper.SuccessWithData(c, result)
}

// @Tags Docker
// @Summary Delete volume
// @Description Delete volume
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param id path string true "Volume ID"
// @Success 200
// @Router /volumes/:host/:id [delete]
func (s *DockerMan) DeleteVolume(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	volumeID := c.Param("id")
	if volumeID == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid volume id", err)
		return
	}

	err = s.deleteVolume(hostID, volumeID)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
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
// @Param page query int true "Page number"
// @Param page_size query int true "Page size"
// @Success 200 {object} model.PageResult
// @Router /networks/:host [get]
func (s *DockerMan) GetNetworks(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	result, err := s.getNetworks(hostID)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}

	helper.SuccessWithData(c, result)
}

// @Tags Docker
// @Summary Create network
// @Description Create network
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param request body model.CreateNetwork true "Create network details"
// @Success 200
// @Router /networks/:host [post]
func (s *DockerMan) CreateNetwork(c *gin.Context) {
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
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Docker
// @Summary Prune network
// @Description Prune network
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Success 200
// @Router /networks/:host/prune [post]
func (s *DockerMan) PruneNetwork(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	err = s.pruneNetwork(hostID)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Docker
// @Summary Get network
// @Description Get network
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param id path string true "Network ID"
// @Success 200 {object} model.Network
// @Router /networks/:host/:id [get]
func (s *DockerMan) GetNetwork(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	networkID := c.Param("id")
	if networkID == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid network id", err)
		return
	}

	result, err := s.getNetwork(hostID, networkID)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}

	helper.SuccessWithData(c, result)
}

// @Tags Docker
// @Summary Delete network
// @Description Delete network
// @Accept json
// @Produce json
// @Param host path int true "Host ID"
// @Param id path string true "Network ID"
// @Success 200
// @Router /networks/:host/:id [delete]
func (s *DockerMan) DeleteNetwork(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	networkID := c.Param("id")
	if networkID == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid network id", err)
		return
	}

	err = s.deleteNetwork(hostID, networkID)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}

	helper.SuccessWithData(c, nil)
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

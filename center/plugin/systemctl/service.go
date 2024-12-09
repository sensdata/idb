package systemctl

import (
	_ "embed"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
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

type SystemCtl struct {
	plugin     plugin.Plugin
	pluginConf plugin.PluginConf
	cmdHelper  *helper.CmdHelper
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

	s.cmdHelper = helper.NewCmdHelper("127.0.0.1", strconv.Itoa(conn.CONFMAN.GetConfig().Port), nil)

	api.API.SetUpPluginRouters(
		"sysctl",
		[]plugin.PluginRoute{
			{Method: "GET", Path: "/info", Handler: s.GetPluginInfo},
			{Method: "GET", Path: "/menu", Handler: s.GetMenu},
			{Method: "GET", Path: "/:host/boot", Handler: s.ServiceBoot},
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
// @Summary run systemctl start/stop/restart service commands
// @Description 运行systemctl的服务启动相关命令
// @Param host path uint true "Host ID"
// @Success 200
// @Router /sysctl/{host}/boot [post]
func (s *SystemCtl) ServiceBoot(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	var req model.ServiceBootReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	err = s.serviceBoot(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithData(c, nil)
}

func (s *SystemCtl) serviceBoot(hostID uint64, req model.ServiceBootReq) error {
	_, err := s.cmdHelper.RunSystemCtl(uint(hostID), req.ServiceName, req.Command)
	if err != nil {
		LOG.Info("service boot err %v", err)
		return err
	}
	return nil
}

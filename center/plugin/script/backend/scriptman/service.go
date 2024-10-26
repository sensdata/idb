package scriptman

import (
	_ "embed"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/gin-gonic/gin"
	"github.com/go-git/go-git/v5"
	"github.com/go-resty/resty/v2"
	"github.com/sensdata/idb/center/core/api"
	"github.com/sensdata/idb/center/core/conn"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/log"
	"github.com/sensdata/idb/core/plugin"
	"github.com/sensdata/idb/core/utils"
	"gopkg.in/yaml.v2"
)

type ScriptMan struct {
	pluginConfig plugin.PluginConfig
	scriptConfig ScriptConfig
	restyClient  *resty.Client
}

var Plugin = ScriptMan{}

var LOG *log.Log

//go:embed plug.yaml
var plugYAML []byte

func (s *ScriptMan) Initialize() {
	fmt.Printf("scriptman init begin")

	//初始化日志模块
	logPath := filepath.Join(constant.CenterLogDir, s.pluginConfig.Plugin.Entry)
	if LOG == nil {
		logger, err := log.InitLogger(logPath, fmt.Sprintf("%s.log", s.pluginConfig.Plugin.Entry))
		if err != nil {
			fmt.Printf("Failed to initialize logger: %v \n", err)
			return
		}
		LOG = logger
	}

	if err := yaml.Unmarshal(plugYAML, &s.pluginConfig); err != nil {
		LOG.Error("Failed to load scriptman yaml: %v", err)
		return
	}

	confPath := filepath.Join(constant.CenterConfDir, s.pluginConfig.Plugin.Entry)
	if _, err := toml.DecodeFile(confPath, &s.scriptConfig); err != nil {
		LOG.Error("Failed to load scriptman toml: %v", err)
		return
	}

	if err := InitDirectories(s.scriptConfig.DataPath); err != nil {
		LOG.Error("Failed to init scriptman data dir: %v", err)
		return
	}

	if err := InitGitRepositories(s.scriptConfig.DataPath); err != nil {
		LOG.Error("Failed to init scriptman git repo: %v", err)
		return
	}

	baseUrl := fmt.Sprintf("http://%s:%d/idb/api", "127.0.0.1", conn.CONFMAN.GetConfig().Port)
	LOG.Info("baseurl: %s", baseUrl)

	s.restyClient = resty.New().
		SetBaseURL(baseUrl).
		SetHeader("Content-Type", "application/json")

	api.API.SetUpPluginRouters(
		"script",
		[]plugin.PluginRoute{
			{Method: "GET", Path: "/info", Handler: s.GetPluginInfo},
			{Method: "GET", Path: "/menu", Handler: s.GetMenu},
		},
	)

	LOG.Info("scriptman init end")
}

func InitDirectories(dataPath string) error {
	globalDir := filepath.Join(dataPath, "global")
	localDir := filepath.Join(dataPath, "local")

	// 检查并创建 global 和 local 目录
	if err := utils.EnsurePaths([]string{globalDir, localDir}); err != nil {
		return err
	}

	LOG.Info("Directories initialized: %s, %s", globalDir, localDir)
	return nil
}

func InitGitRepositories(dataPath string) error {
	globalDir := filepath.Join(dataPath, "global")
	localDir := filepath.Join(dataPath, "local")

	// 初始化 global 仓库
	if err := InitGitRepo(globalDir); err != nil {
		LOG.Error("Failed to init global repo %v", err)
		return err
	}

	// 初始化 local 仓库
	if err := InitGitRepo(localDir); err != nil {
		LOG.Error("Failed to init local repo %v", err)
		return err
	}

	LOG.Info("Scriptman repos initialized")
	return nil
}

// 检查并初始化为 Git 仓库
func InitGitRepo(path string) error {
	_, err := git.PlainOpen(path)
	if err == git.ErrRepositoryNotExists {
		LOG.Info("Initializing Git repository at: %s", path)
		_, err := git.PlainInit(path, false) // 初始化非裸仓库
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

func (s *ScriptMan) Release() {

}

// @Tags File
// @Summary Get plugin info
// @Description Get plugin information
// @Accept json
// @Produce json
// @Success 200 {object} plugin.PluginInfo
// @Router /files/info [get]
func (s *ScriptMan) GetPluginInfo(c *gin.Context) {
	pluginInfo, err := s.getPluginInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"type": "info", "payload": pluginInfo})
}

// @Tags File
// @Summary Get plugin menu
// @Description Get plugin menu items
// @Accept json
// @Produce json
// @Success 200 {array} plugin.MenuItem
// @Router /files/menu [get]
func (s *ScriptMan) GetMenu(c *gin.Context) {
	menuItems, err := s.getMenus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"type": "menu", "payload": menuItems})
}

func (s *ScriptMan) getPluginInfo() (plugin.PluginInfo, error) {
	return s.pluginConfig.Plugin, nil
}

func (s *ScriptMan) getMenus() ([]plugin.MenuItem, error) {
	return s.pluginConfig.Menu, nil
}

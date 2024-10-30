package scriptman

import (
	_ "embed"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/BurntSushi/toml"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/sensdata/idb/center/core/api"
	"github.com/sensdata/idb/center/core/conn"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/helper"
	"github.com/sensdata/idb/core/log"
	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/plugin"
	"gopkg.in/yaml.v2"
)

type ScriptMan struct {
	pluginConfig plugin.PluginConfig
	scriptConfig *model.Script
	restyClient  *resty.Client
}

var Plugin = ScriptMan{}

var LOG *log.Log

//go:embed plug.yaml
var plugYAML []byte

func (s *ScriptMan) Initialize() {
	fmt.Printf("scriptman init begin \n")

	confPath := filepath.Join(constant.CenterConfDir, "script", "script.toml")
	// 检查配置文件的目录是否存在
	if err := os.MkdirAll(filepath.Dir(confPath), os.ModePerm); err != nil {
		fmt.Printf("Failed to create directory: %v \n", err)
		return
	}
	// 检查配置文件是否存在
	if _, err := os.Stat(confPath); os.IsNotExist(err) {
		// 创建配置文件并写入默认内容
		defaultConfig := "[script]\ndata_path = /var/lib/idb/data/script\nlog_path = /var/lib/idb\n"
		if err := os.WriteFile(confPath, []byte(defaultConfig), 0644); err != nil {
			fmt.Printf("Failed to create script toml: %v \n", err)
			return
		}
	}
	if _, err := toml.DecodeFile(confPath, &s.scriptConfig); err != nil {
		fmt.Printf("Failed to load script toml: %v \n", err)
		return
	}

	//初始化日志模块
	if LOG == nil {
		logger, err := log.InitLogger(s.scriptConfig.Script.LogPath, "script.log")
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

	baseUrl := fmt.Sprintf("http://%s:%d/api/v1", "127.0.0.1", conn.CONFMAN.GetConfig().Port)
	LOG.Info("baseurl: %s", baseUrl)

	s.restyClient = resty.New().
		SetBaseURL(baseUrl).
		SetHeader("Content-Type", "application/json")

	api.API.SetUpPluginRouters(
		"scripts",
		[]plugin.PluginRoute{
			{Method: "GET", Path: "/info", Handler: s.GetPluginInfo},
			{Method: "GET", Path: "/menu", Handler: s.GetMenu},
			{Method: "GET", Path: "", Handler: s.GetScriptList},
			{Method: "GET", Path: "/detail", Handler: s.GetScriptDetail},
			{Method: "POST", Path: "", Handler: s.Create},
			{Method: "PUT", Path: "", Handler: s.Update},
			{Method: "DELETE", Path: "", Handler: s.Delete},
			{Method: "PUT", Path: "/restore", Handler: s.Restore},
			{Method: "GET", Path: "/log", Handler: s.GetScriptLog},
			{Method: "GET", Path: "/diff", Handler: s.GetScriptDiff},
		},
	)

	LOG.Info("scriptman init end")
}

func (s *ScriptMan) Release() {

}

// @Tags File
// @Summary Get plugin info
// @Description Get plugin information
// @Accept json
// @Produce json
// @Success 200 {object} plugin.PluginInfo
// @Router /scripts/info [get]
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
// @Router /scripts/menu [get]
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

// @Tags Script
// @Summary List scripts
// @Description Get list of scripts in a directory
// @Accept json
// @Produce json
// @Param host_id query uint true "Host ID"
// @Param type query string true "Type (options: 'global', 'local')"
// @Param category query string false "Category (directory under 'global' or 'local')"
// @Param page query uint true "Page"
// @Param page_size query uint true "Page size"
// @Success 200 {array} model.PageResult
// @Router /scripts [get]
func (s *ScriptMan) GetScriptList(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Query("host_id"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host_id", err)
		return
	}

	scriptType := c.Query("type")
	if scriptType == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid type", err)
		return
	}
	if scriptType != "global" && scriptType != "local" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid type", err)
		return
	}

	category := c.Query("category")

	page, err := strconv.ParseInt(c.Query("page"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid page", err)
		return
	}

	pageSize, err := strconv.ParseInt(c.Query("page_size"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid page_size", err)
		return
	}

	req := model.QueryScript{
		HostID:   uint(hostID),
		Type:     scriptType,
		Category: category,
		Page:     int(page),
		PageSize: int(pageSize),
	}

	scripts, err := s.getScriptList(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}

	helper.SuccessWithData(c, scripts)
}

// @Tags Script
// @Summary Get script detail
// @Description Get detail of a script file
// @Accept json
// @Produce json
// @Param host_id query uint true "Host ID"
// @Param type query string true "Type (options: 'global', 'local')"
// @Param category query string false "Category (directory under 'global' or 'local')"
// @Param name query string true "Script file name"
// @Success 200 {array} model.GitFile
// @Router /scripts/detail [get]
func (s *ScriptMan) GetScriptDetail(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Query("host_id"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host_id", err)
		return
	}

	scriptType := c.Query("type")
	if scriptType == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid type", err)
		return
	}
	if scriptType != "global" && scriptType != "local" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid type", err)
		return
	}

	category := c.Query("category")

	name := c.Query("name")
	if name == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid name", err)
		return
	}

	req := model.GetScript{
		HostID:   uint(hostID),
		Type:     scriptType,
		Category: category,
		Name:     name,
	}

	detail, err := s.getScriptDetail(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}

	helper.SuccessWithData(c, detail)
}

// @Tags Script
// @Summary Create script file or category
// @Description Create a new script file or category
// @Accept json
// @Produce json
// @Param request body model.CreateScript true "Script file creation details"
// @Success 200
// @Router /scripts [post]
func (s *ScriptMan) Create(c *gin.Context) {
	var req model.CreateScript
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	err := s.create(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Script
// @Summary Update script file content
// @Description Update the content of a script file
// @Accept json
// @Produce json
// @Param request body model.UpdateScript true "Script file edit details"
// @Success 200
// @Router /scripts [put]
func (s *ScriptMan) Update(c *gin.Context) {
	var req model.UpdateScript
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	err := s.update(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Script
// @Summary Delete script file
// @Description Delete  a script file
// @Accept json
// @Produce json
// @Param host_id query uint true "Host ID"
// @Param type query string true "Type (options: 'global', 'local')"
// @Param category query string false "Category (directory under 'global' or 'local')"
// @Param name query string true "File name"
// @Success 200
// @Router /scripts [delete]
func (s *ScriptMan) Delete(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Query("host_id"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host_id", err)
		return
	}

	scriptType := c.Query("type")
	if scriptType == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid type", err)
		return
	}
	if scriptType != "global" && scriptType != "local" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid type", err)
		return
	}

	category := c.Query("category")

	name := c.Query("name")
	if name == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid name", err)
		return
	}

	req := model.DeleteScript{
		HostID:   uint(hostID),
		Type:     scriptType,
		Category: category,
		Name:     name,
	}

	err = s.delete(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Script
// @Summary Restore script file
// @Description Restore script file to specified version
// @Accept json
// @Produce json
// @Param request body model.RestoreScript true "Script file restore details"
// @Success 200
// @Router /scripts/restore [put]
func (s *ScriptMan) Restore(c *gin.Context) {
	var req model.RestoreScript
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	err := s.restore(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Script
// @Summary Get script histories
// @Description Get histories of a script file
// @Accept json
// @Produce json
// @Param host_id query uint true "Host ID"
// @Param type query string true "Type (options: 'global', 'local')"
// @Param category query string false "Category (directory under 'global' or 'local')"
// @Param name query string true "Script file name"
// @Param page query uint true "Page"
// @Param page_size query uint true "Page size"
// @Success 200 {array} model.PageResult
// @Router /scripts/log [get]
func (s *ScriptMan) GetScriptLog(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Query("host_id"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host_id", err)
		return
	}

	scriptType := c.Query("type")
	if scriptType == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid type", err)
		return
	}
	if scriptType != "global" && scriptType != "local" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid type", err)
		return
	}

	category := c.Query("category")

	name := c.Query("name")
	if name == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid name", err)
		return
	}

	page, err := strconv.ParseInt(c.Query("page"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid page", err)
		return
	}

	pageSize, err := strconv.ParseInt(c.Query("page_size"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid page_size", err)
		return
	}

	req := model.ScriptLog{
		HostID:   uint(hostID),
		Type:     scriptType,
		Category: category,
		Name:     name,
		Page:     int(page),
		PageSize: int(pageSize),
	}

	logs, err := s.getScriptLog(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}

	helper.SuccessWithData(c, logs)
}

// @Tags Script
// @Summary Get script diff
// @Description Get script diff compare to specfied version
// @Accept json
// @Produce json
// @Param host_id query uint true "Host ID"
// @Param type query string true "Type (options: 'global', 'local')"
// @Param category query string false "Category (directory under 'global' or 'local')"
// @Param name query string true "Script file name"
// @Param commit query string true "Commit hash"
// @Success 200 {array} model.GitFile
// @Router /scripts/diff [get]
func (s *ScriptMan) GetScriptDiff(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Query("host_id"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host_id", err)
		return
	}

	scriptType := c.Query("type")
	if scriptType == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid type", err)
		return
	}
	if scriptType != "global" && scriptType != "local" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid type", err)
		return
	}

	category := c.Query("category")

	name := c.Query("name")
	if name == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid name", err)
		return
	}

	commitHash := c.Query("commit")
	if commitHash == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid commit hash", err)
		return
	}

	req := model.ScriptDiff{
		HostID:     uint(hostID),
		Type:       scriptType,
		Category:   category,
		Name:       name,
		CommitHash: commitHash,
	}

	diff, err := s.getScriptDiff(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}

	helper.SuccessWithData(c, diff)
}

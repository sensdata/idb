package crontab

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

type CronTab struct {
	plugin       plugin.Plugin
	pluginConf   plugin.PluginConf
	form         model.Form
	templateForm model.ServiceForm
	restyClient  *resty.Client
}

var LOG *log.Log

//go:embed plug.yaml
var plugYAML []byte

//go:embed conf.yaml
var confYAML []byte

//go:embed form.yaml
var formYaml []byte

//go:embed template.crontab
var templateService []byte

func (s *CronTab) Initialize() {
	global.LOG.Info("crontab init begin \n")

	// 解析plugYAML
	if err := yaml.Unmarshal(plugYAML, &s.plugin); err != nil {
		global.LOG.Error("Failed to load info: %v", err)
		return
	}

	// 解析formYaml
	if err := yaml.Unmarshal(formYaml, &s.form); err != nil {
		global.LOG.Error("Failed to load form: %v", err)
		return
	}

	// 由templateService解析出模板的templateServiceForm
	var err error
	s.templateForm, err = parseConfBytesToServiceForm(templateService, s.form.Fields)
	if err != nil {
		global.LOG.Error("Failed to parse template: %v", err)
		return
	}

	confPath := filepath.Join(constant.CenterConfDir, "crontab", "conf.yaml")
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

	global.LOG.Info("crontab conf: %v", s.pluginConf)

	//初始化日志模块
	if LOG == nil {
		logger, err := log.InitLogger(s.pluginConf.Items.LogDir, "crontab.log")
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
		"crontab",
		[]plugin.PluginRoute{
			{Method: "GET", Path: "/info", Handler: s.GetPluginInfo},
			{Method: "GET", Path: "/menu", Handler: s.GetMenu},
			{Method: "GET", Path: "", Handler: s.GetConfList},
			{Method: "GET", Path: "/raw", Handler: s.GetContent},     // 源文模式获取
			{Method: "POST", Path: "/raw", Handler: s.CreateContent}, // 源文模式创建
			{Method: "PUT", Path: "/raw", Handler: s.UpdateContent},  // 源文模式更新
			{Method: "GET", Path: "/form", Handler: s.GetForm},       // 表单模式获取
			{Method: "POST", Path: "/form", Handler: s.CreateForm},   // 表单模式创建
			{Method: "PUT", Path: "/form", Handler: s.UpdateForm},    // 表单模式更新
			{Method: "DELETE", Path: "", Handler: s.Delete},
			{Method: "PUT", Path: "/restore", Handler: s.Restore},
			{Method: "GET", Path: "/log", Handler: s.GetConfLog},
			{Method: "GET", Path: "/diff", Handler: s.GetConfDiff},
			{Method: "POST", Path: "/action", Handler: s.ConfAction},
		},
	)

	global.LOG.Info("crontab init end")
}

func (s *CronTab) Release() {

}

// @Tags Crontab
// @Summary Get plugin info
// @Description Get plugin information
// @Accept json
// @Produce json
// @Success 200 {object} plugin.PluginInfo
// @Router /crontab/info [get]
func (s *CronTab) GetPluginInfo(c *gin.Context) {
	pluginInfo, err := s.getPluginInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"type": "info", "payload": pluginInfo})
}

// @Tags Crontab
// @Summary Get plugin menu
// @Description Get plugin menu items
// @Accept json
// @Produce json
// @Success 200 {array} plugin.MenuItem
// @Router /crontab/menu [get]
func (s *CronTab) GetMenu(c *gin.Context) {
	menuItems, err := s.getMenus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"type": "menu", "payload": menuItems})
}

func (s *CronTab) getPluginInfo() (plugin.PluginInfo, error) {
	return s.plugin.Info, nil
}

func (s *CronTab) getMenus() ([]plugin.MenuItem, error) {
	return s.plugin.Menu, nil
}

// @Tags Crontab
// @Summary List crontab conf files
// @Description Get custom crontab conf file list in work dir
// @Accept json
// @Produce json
// @Param host_id query uint true "Host ID"
// @Param type query string true "Type (options: 'global', 'local')"
// @Param category query string false "Category (directory under 'global' or 'local')"
// @Param page query uint true "Page"
// @Param page_size query uint true "Page size"
// @Success 200 {object} model.PageResult
// @Router /crontab [get]
func (s *CronTab) GetConfList(c *gin.Context) {
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

	req := model.QueryGitFile{
		HostID:   uint(hostID),
		Type:     scriptType,
		Category: category,
		Page:     int(page),
		PageSize: int(pageSize),
	}

	services, err := s.getConfList(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}

	helper.SuccessWithData(c, services)
}

// @Tags Crontab
// @Summary Create conf file
// @Description Create a new conf file
// @Accept json
// @Produce json
// @Param request body model.CreateGitFile true "Conf file creation details"
// @Success 200
// @Router /crontab/raw [post]
func (s *CronTab) CreateContent(c *gin.Context) {
	var req model.CreateGitFile
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	err := s.create(req, ".crontab")
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Crontab
// @Summary Get conf file content
// @Description Get content of a conf file
// @Accept json
// @Produce json
// @Param host_id query uint true "Host ID"
// @Param type query string true "Type (options: 'global', 'local')"
// @Param category query string false "Category (directory under 'global' or 'local')"
// @Param name query string true "Conf file name"
// @Success 200 {string} string
// @Router /crontab/raw [get]
func (s *CronTab) GetContent(c *gin.Context) {
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

	req := model.GetGitFileDetail{
		HostID:   uint(hostID),
		Type:     scriptType,
		Category: category,
		Name:     name,
	}

	detail, err := s.getContent(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}

	helper.SuccessWithData(c, detail.Content)
}

// @Tags Crontab
// @Summary Save conf file content
// @Description Save the content of a conf file
// @Accept json
// @Produce json
// @Param request body model.UpdateGitFile true "Conf file edit details"
// @Success 200
// @Router /crontab/raw [put]
func (s *CronTab) UpdateContent(c *gin.Context) {
	var req model.UpdateGitFile
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

// @Tags Crontab
// @Summary Get conf file in form mode
// @Description Get details of a conf file in form mode.
// @Accept json
// @Produce json
// @Param host_id query uint true "Host ID"
// @Param type query string true "Type (options: 'global', 'local')"
// @Param category query string false "Category (directory under 'global' or 'local')"
// @Param name query string false "Conf file name. If this parameter is left empty, return template data."
// @Success 200 {object} model.ServiceForm
// @Router /crontab/form [get]
func (s *CronTab) GetForm(c *gin.Context) {
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

	req := model.GetGitFileDetail{
		HostID:   uint(hostID),
		Type:     scriptType,
		Category: category,
		Name:     name,
	}

	detail, err := s.getForm(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}

	helper.SuccessWithData(c, detail)
}

// @Tags Crontab
// @Summary Create conf file in form mode
// @Description Create a new conf file in form mode
// @Accept json
// @Produce json
// @Param request body model.CreateServiceForm true "Form details"
// @Success 200
// @Router /crontab/form [post]
func (s *CronTab) CreateForm(c *gin.Context) {
	var req model.CreateServiceForm
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	err := s.createForm(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Crontab
// @Summary Save conf file in form mode
// @Description Save the details of a conf file in form mode
// @Accept json
// @Produce json
// @Param request body model.UpdateServiceForm true "Conf file edit details"
// @Success 200
// @Router /crontab/form [put]
func (s *CronTab) UpdateForm(c *gin.Context) {
	var req model.UpdateServiceForm
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	err := s.updateForm(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Crontab
// @Summary Delete conf file
// @Description Delete a conf file
// @Accept json
// @Produce json
// @Param host_id query uint true "Host ID"
// @Param type query string true "Type (options: 'global', 'local')"
// @Param category query string false "Category (directory under 'global' or 'local')"
// @Param name query string true "File name"
// @Success 200
// @Router /crontab [delete]
func (s *CronTab) Delete(c *gin.Context) {
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

	req := model.DeleteGitFile{
		HostID:   uint(hostID),
		Type:     scriptType,
		Category: category,
		Name:     name,
	}

	err = s.delete(req, ".crontab")
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Crontab
// @Summary Restore conf file
// @Description Restore conf file to specified version
// @Accept json
// @Produce json
// @Param request body model.RestoreGitFile true "Conf file restore details"
// @Success 200
// @Router /crontab/restore [put]
func (s *CronTab) Restore(c *gin.Context) {
	var req model.RestoreGitFile
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

// @Tags Crontab
// @Summary Get conf histories
// @Description Get histories of a conf file
// @Accept json
// @Produce json
// @Param host_id query uint true "Host ID"
// @Param type query string true "Type (options: 'global', 'local')"
// @Param category query string false "Category (directory under 'global' or 'local')"
// @Param name query string true "Conf file name"
// @Param page query uint true "Page"
// @Param page_size query uint true "Page size"
// @Success 200 {object} model.PageResult
// @Router /crontab/log [get]
func (s *CronTab) GetConfLog(c *gin.Context) {
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

	req := model.GitFileLog{
		HostID:   uint(hostID),
		Type:     scriptType,
		Category: category,
		Name:     name,
		Page:     int(page),
		PageSize: int(pageSize),
	}

	logs, err := s.getConfLog(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}

	helper.SuccessWithData(c, logs)
}

// @Tags Crontab
// @Summary Get conf diff
// @Description Get conf diff compare to specfied version
// @Accept json
// @Produce json
// @Param host_id query uint true "Host ID"
// @Param type query string true "Type (options: 'global', 'local')"
// @Param category query string false "Category (directory under 'global' or 'local')"
// @Param name query string true "Conf file name"
// @Param commit query string true "Commit hash"
// @Success 200 {string} string
// @Router /crontab/diff [get]
func (s *CronTab) GetConfDiff(c *gin.Context) {
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

	req := model.GitFileDiff{
		HostID:     uint(hostID),
		Type:       scriptType,
		Category:   category,
		Name:       name,
		CommitHash: commitHash,
	}

	diff, err := s.getConfDiff(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}

	helper.SuccessWithData(c, diff)
}

// @Tags Crontab
// @Summary Execute conf actions
// @Description Execute conf actions
// @Accept json
// @Produce json
// @Param request body model.ConfAction true "Conf action details"
// @Success 200
// @Router /crontab/action [post]
func (s *CronTab) ConfAction(c *gin.Context) {
	var req model.ServiceAction
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	err := s.confAction(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithData(c, nil)
}

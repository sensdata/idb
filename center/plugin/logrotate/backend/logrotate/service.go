package logrotate

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
	"github.com/sensdata/idb/center/db/repo"
	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/helper"
	"github.com/sensdata/idb/core/log"
	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/plugin"
	"gopkg.in/yaml.v2"
)

type LogRotate struct {
	plugin       plugin.Plugin
	pluginConf   plugin.PluginConf
	form         model.Form
	templateForm model.ServiceForm
	restyClient  *resty.Client
	hostRepo     repo.IHostRepo
}

var LOG *log.Log

//go:embed plug.yaml
var plugYAML []byte

//go:embed conf.yaml
var confYAML []byte

//go:embed form.yaml
var formYaml []byte

//go:embed template.logrotate
var templateService []byte

func (s *LogRotate) Initialize() {
	global.LOG.Info("logrotate init begin \n")

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

	confPath := filepath.Join(constant.CenterConfDir, "logrotate", "conf.yaml")
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

	global.LOG.Info("logrotate conf: %v", s.pluginConf)

	//初始化日志模块
	if LOG == nil {
		logger, err := log.InitLogger(s.pluginConf.Items.LogDir, "logrotate.log")
		if err != nil {
			global.LOG.Error("Failed to initialize logger: %v \n", err)
			return
		}
		LOG = logger
	}

	api.API.SetUpPluginRouters(
		"logrotate",
		[]plugin.PluginRoute{
			{Method: "GET", Path: "/info", Handler: s.GetPluginInfo},
			{Method: "GET", Path: "/menu", Handler: s.GetMenu},
			{Method: "GET", Path: "/:host/category", Handler: s.GetCategories},
			{Method: "POST", Path: "/:host/category", Handler: s.CreateCategory},
			{Method: "PUT", Path: "/:host/category", Handler: s.UpdateCategory},
			{Method: "DELETE", Path: "/:host/category", Handler: s.DeleteCategory},
			{Method: "GET", Path: "/:host", Handler: s.GetConfList},
			{Method: "GET", Path: "/:host/raw", Handler: s.GetContent},     // 源文模式获取
			{Method: "POST", Path: "/:host/raw", Handler: s.CreateContent}, // 源文模式创建
			{Method: "PUT", Path: "/:host/raw", Handler: s.UpdateContent},  // 源文模式更新
			{Method: "GET", Path: "/:host/form", Handler: s.GetForm},       // 表单模式获取
			{Method: "POST", Path: "/:host/form", Handler: s.CreateForm},   // 表单模式创建
			{Method: "PUT", Path: "/:host/form", Handler: s.UpdateForm},    // 表单模式更新
			{Method: "DELETE", Path: "/:host", Handler: s.Delete},
			{Method: "PUT", Path: "/:host/restore", Handler: s.Restore},
			{Method: "GET", Path: "/:host/history", Handler: s.GetConfLog},
			{Method: "GET", Path: "/:host/diff", Handler: s.GetConfDiff},
			{Method: "POST", Path: "/:host/sync", Handler: s.SyncGlobal},
			{Method: "POST", Path: "/:host/activate", Handler: s.ConfActivate},
			{Method: "POST", Path: "/:host/operate", Handler: s.ConfOperate},
		},
	)

	global.LOG.Info("logrotate init end")
}

func (s *LogRotate) Start() {
	global.LOG.Info("logrotate start")
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

	// 初始化 host 模块
	s.hostRepo = repo.NewHostRepo()
}

func (s *LogRotate) Release() {

}

// @Tags Logrotate
// @Summary Get plugin info
// @Description Get plugin information
// @Accept json
// @Produce json
// @Success 200 {object} plugin.PluginInfo
// @Router /logrotate/info [get]
func (s *LogRotate) GetPluginInfo(c *gin.Context) {
	pluginInfo, err := s.getPluginInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"type": "info", "payload": pluginInfo})
}

// @Tags Logrotate
// @Summary Get plugin menu
// @Description Get plugin menu items
// @Accept json
// @Produce json
// @Success 200 {array} plugin.MenuItem
// @Router /logrotate/menu [get]
func (s *LogRotate) GetMenu(c *gin.Context) {
	menuItems, err := s.getMenus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"type": "menu", "payload": menuItems})
}

func (s *LogRotate) getPluginInfo() (plugin.PluginInfo, error) {
	return s.plugin.Info, nil
}

func (s *LogRotate) getMenus() ([]plugin.MenuItem, error) {
	return s.plugin.Menu, nil
}

// @Tags Logrotate
// @Summary List category
// @Description List category
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param type query string true "Type (options: 'global', 'local')"
// @Param page query uint true "Page"
// @Param page_size query uint true "Page size"
// @Success 200 {object} model.PageResult
// @Router /logrotate/{host}/category [get]
func (s *LogRotate) GetCategories(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
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
		Type:     scriptType,
		Category: "",
		Page:     int(page),
		PageSize: int(pageSize),
	}

	categories, err := s.getCategories(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	helper.SuccessWithData(c, categories)
}

// @Tags Logrotate
// @Summary Create category
// @Description Create category
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param request body model.CreateGitCategory true "Category creation details"
// @Success 200
// @Router /logrotate/{host}/category [post]
func (s *LogRotate) CreateCategory(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	var req model.CreateGitCategory
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	err = s.createCategory(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Logrotate
// @Summary Update category
// @Description Update category
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param request body model.UpdateGitCategory true "category edit details"
// @Success 200
// @Router /logrotate/{host}/category [put]
func (s *LogRotate) UpdateCategory(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	var req model.UpdateGitCategory
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	err = s.updateCategory(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Logrotate
// @Summary Delete category
// @Description Delete category
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param type query string true "Type (options: 'global', 'local')"
// @Param category query string true "Category (directory under 'global' or 'local')"
// @Success 200
// @Router /logrotate/{host}/category [delete]
func (s *LogRotate) DeleteCategory(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
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
	if category == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid category", err)
		return
	}

	req := model.DeleteGitCategory{
		Type:     scriptType,
		Category: category,
	}

	err = s.deleteCategory(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Logrotate
// @Summary List logrotate conf files
// @Description Get custom logrotate conf file list in work dir
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param type query string true "Type (options: 'global', 'local')"
// @Param category query string true "Category (directory under 'global' or 'local')"
// @Param page query uint true "Page"
// @Param page_size query uint true "Page size"
// @Success 200 {object} model.PageResult
// @Router /logrotate/{host} [get]
func (s *LogRotate) GetConfList(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
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
	if category == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid category", err)
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

	req := model.QueryGitFile{
		Type:     scriptType,
		Category: category,
		Page:     int(page),
		PageSize: int(pageSize),
	}

	services, err := s.getConfList(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	helper.SuccessWithData(c, services)
}

// @Tags Logrotate
// @Summary Create conf file
// @Description Create a new conf file
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param request body model.CreateGitFile true "Conf file creation details"
// @Success 200
// @Router /logrotate/{host}/raw [post]
func (s *LogRotate) CreateContent(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	var req model.CreateGitFile
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	err = s.create(hostID, req, ".logrotate")
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Logrotate
// @Summary Get conf file content
// @Description Get content of a conf file
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param type query string true "Type (options: 'global', 'local')"
// @Param category query string true "Category (directory under 'global' or 'local')"
// @Param name query string true "Conf file name"
// @Success 200 {string} string
// @Router /logrotate/{host}/raw [get]
func (s *LogRotate) GetContent(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
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
	if category == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid category", err)
		return
	}

	name := c.Query("name")
	if name == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid name", err)
		return
	}

	req := model.GetGitFileDetail{
		Type:     scriptType,
		Category: category,
		Name:     name,
	}

	detail, err := s.getContent(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	helper.SuccessWithData(c, detail.Content)
}

// @Tags Logrotate
// @Summary Save conf file content
// @Description Save the content of a conf file
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param request body model.UpdateGitFile true "Conf file edit details"
// @Success 200
// @Router /logrotate/{host}/raw [put]
func (s *LogRotate) UpdateContent(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	var req model.UpdateGitFile
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	err = s.update(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Logrotate
// @Summary Get conf file in form mode
// @Description Get details of a conf file in form mode.
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param type query string true "Type (options: 'global', 'local')"
// @Param category query string true "Category (directory under 'global' or 'local')"
// @Param name query string false "Conf file name. If this parameter is left empty, return template data."
// @Success 200 {object} model.ServiceForm
// @Router /logrotate/{host}/form [get]
func (s *LogRotate) GetForm(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
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
	if category == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid category", err)
		return
	}

	name := c.Query("name")

	req := model.GetGitFileDetail{
		Type:     scriptType,
		Category: category,
		Name:     name,
	}

	detail, err := s.getForm(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	helper.SuccessWithData(c, detail)
}

// @Tags Logrotate
// @Summary Create conf file in form mode
// @Description Create a new conf file in form mode
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param request body model.CreateServiceForm true "Form details"
// @Success 200
// @Router /logrotate/{host}/form [post]
func (s *LogRotate) CreateForm(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	var req model.CreateServiceForm
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	err = s.createForm(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Logrotate
// @Summary Save conf file in form mode
// @Description Save the details of a conf file in form mode
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param request body model.UpdateServiceForm true "Conf file edit details"
// @Success 200
// @Router /logrotate/{host}/form [put]
func (s *LogRotate) UpdateForm(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	var req model.UpdateServiceForm
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	err = s.updateForm(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Logrotate
// @Summary Delete conf file
// @Description Delete a conf file
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param type query string true "Type (options: 'global', 'local')"
// @Param category query string true "Category (directory under 'global' or 'local')"
// @Param name query string true "File name"
// @Success 200
// @Router /logrotate/{host} [delete]
func (s *LogRotate) Delete(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
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
	if category == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid category", err)
		return
	}

	name := c.Query("name")
	if name == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid name", err)
		return
	}

	req := model.DeleteGitFile{
		Type:     scriptType,
		Category: category,
		Name:     name,
	}

	err = s.delete(hostID, req, ".logrotate")
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Logrotate
// @Summary Restore conf file
// @Description Restore conf file to specified version
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param request body model.RestoreGitFile true "Conf file restore details"
// @Success 200
// @Router /logrotate/{host}/restore [put]
func (s *LogRotate) Restore(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	var req model.RestoreGitFile
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	err = s.restore(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Logrotate
// @Summary Get conf histories
// @Description Get histories of a conf file
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param type query string true "Type (options: 'global', 'local')"
// @Param category query string true "Category (directory under 'global' or 'local')"
// @Param name query string true "Conf file name"
// @Param page query uint true "Page"
// @Param page_size query uint true "Page size"
// @Success 200 {object} model.PageResult
// @Router /logrotate/{host}/history [get]
func (s *LogRotate) GetConfLog(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
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
	if category == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid category", err)
		return
	}

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
		Type:     scriptType,
		Category: category,
		Name:     name,
		Page:     int(page),
		PageSize: int(pageSize),
	}

	logs, err := s.getConfLog(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	helper.SuccessWithData(c, logs)
}

// @Tags Logrotate
// @Summary Get conf diff
// @Description Get conf diff compare to specfied version
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param type query string true "Type (options: 'global', 'local')"
// @Param category query string true "Category (directory under 'global' or 'local')"
// @Param name query string true "Conf file name"
// @Param commit query string true "Commit hash"
// @Success 200 {string} string
// @Router /logrotate/{host}/diff [get]
func (s *LogRotate) GetConfDiff(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
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
	if category == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid category", err)
		return
	}

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
		Type:       scriptType,
		Category:   category,
		Name:       name,
		CommitHash: commitHash,
	}

	diff, err := s.getConfDiff(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	helper.SuccessWithData(c, diff)
}

// @Tags Logrotate
// @Summary Sync global repository to specified host
// @Description Sync global repository to specified host
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Success 200
// @Router /logrotate/{host}/sync [post]
func (s *LogRotate) SyncGlobal(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	err = s.syncGlobal(uint(hostID))
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Logrotate
// @Summary Activate or deactivate conf
// @Description Create a symlink in the logrotate configuration directory to activate the specified configuration, or remove the symlink to deactivate it.
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param request body model.ServiceActivate true "Conf action details"
// @Success 200
// @Router /logrotate/{host}/activate [post]
func (s *LogRotate) ConfActivate(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	var req model.ServiceActivate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	err = s.confActivate(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Logrotate
// @Summary Test or force execute logrotate configuration
// @Description Test (dry-run) or force execute a specified logrotate configuration for the given host, returning the operation result.
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param request body model.LogrotateOperate true "Logrotate operation details"
// @Success 200 {object} model.ServiceOperateResult
// @Router /logrotate/{host}/operate [post]
func (s *LogRotate) ConfOperate(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	var req model.LogrotateOperate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	result, err := s.confOperate(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}
	helper.SuccessWithData(c, result)
}

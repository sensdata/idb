package certificate

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

type CertificateMan struct {
	plugin      plugin.Plugin
	pluginConf  plugin.PluginConf
	restyClient *resty.Client
}

var LOG *log.Log

//go:embed plug.yaml
var plugYAML []byte

//go:embed conf.yaml
var confYAML []byte

func (s *CertificateMan) Initialize() {
	global.LOG.Info("certificateman init begin \n")

	if err := yaml.Unmarshal(plugYAML, &s.plugin); err != nil {
		global.LOG.Error("Failed to load info: %v", err)
		return
	}

	confPath := filepath.Join(constant.CenterConfDir, "certificate", "conf.yaml")
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
		logger, err := log.InitLogger(s.pluginConf.Items.LogDir, "certificate.log")
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
		"certificates",
		[]plugin.PluginRoute{
			{Method: "GET", Path: "/info", Handler: s.GetPluginInfo},
			{Method: "GET", Path: "/menu", Handler: s.GetMenu},
			{Method: "GET", Path: "", Handler: s.GetCertificates},
			{Method: "GET", Path: "/detail", Handler: s.GetScriptDetail},
			{Method: "POST", Path: "", Handler: s.Create},
			{Method: "DELETE", Path: "", Handler: s.Delete},
		},
	)

	global.LOG.Info("certificateman init end")
}

func (s *CertificateMan) Release() {

}

// @Tags Script
// @Summary Get plugin info
// @Description Get plugin information
// @Accept json
// @Produce json
// @Success 200 {object} plugin.PluginInfo
// @Router /scripts/info [get]
func (s *CertificateMan) GetPluginInfo(c *gin.Context) {
	pluginInfo, err := s.getPluginInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"type": "info", "payload": pluginInfo})
}

// @Tags Script
// @Summary Get plugin menu
// @Description Get plugin menu items
// @Accept json
// @Produce json
// @Success 200 {array} plugin.MenuItem
// @Router /scripts/menu [get]
func (s *CertificateMan) GetMenu(c *gin.Context) {
	menuItems, err := s.getMenus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"type": "menu", "payload": menuItems})
}

func (s *CertificateMan) getPluginInfo() (plugin.PluginInfo, error) {
	return s.plugin.Info, nil
}

func (s *CertificateMan) getMenus() ([]plugin.MenuItem, error) {
	return s.plugin.Menu, nil
}

// @Tags Script
// @Summary List scripts
// @Description Get list of scripts in a directory
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param type query string true "Type (options: 'global', 'local')"
// @Param category query string false "Category (directory under 'global' or 'local')"
// @Param page query uint true "Page"
// @Param page_size query uint true "Page size"
// @Success 200 {object} model.PageResult
// @Router /scripts/{host} [get]
func (s *CertificateMan) GetCertificates(c *gin.Context) {
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

	scripts, err := s.getScriptList(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}

	helper.SuccessWithData(c, scripts)
}

// @Tags Script
// @Summary Get certificate detail
// @Description Get detail of a certificate file
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param type query string true "Type (options: 'global', 'local')"
// @Param category query string false "Category (directory under 'global' or 'local')"
// @Param name query string true "Script file name"
// @Success 200 {object} model.GitFile
// @Router /scripts/{host}/detail [get]
func (s *CertificateMan) GetScriptDetail(c *gin.Context) {
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

	detail, err := s.getScriptDetail(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}

	helper.SuccessWithData(c, detail)
}

// @Tags Script
// @Summary Create certificate file or category
// @Description Create a new certificate file or category
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param request body model.CreateGitFile true "Script file creation details"
// @Success 200
// @Router /scripts [post]
func (s *CertificateMan) Create(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	var req model.CreateGitFile
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	err = s.create(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Script
// @Summary Update certificate file content
// @Description Update the content of a certificate file
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param request body model.UpdateGitFile true "Script file edit details"
// @Success 200
// @Router /scripts [put]
func (s *CertificateMan) Update(c *gin.Context) {
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
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Script
// @Summary Delete certificate file
// @Description Delete  a certificate file
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param type query string true "Type (options: 'global', 'local')"
// @Param category query string false "Category (directory under 'global' or 'local')"
// @Param name query string true "File name"
// @Success 200
// @Router /scripts [delete]
func (s *CertificateMan) Delete(c *gin.Context) {
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

	err = s.delete(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Script
// @Summary Restore certificate file
// @Description Restore certificate file to specified version
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param request body model.RestoreGitFile true "Script file restore details"
// @Success 200
// @Router /scripts/{host}/restore [put]
func (s *CertificateMan) Restore(c *gin.Context) {
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
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Script
// @Summary Get certificate histories
// @Description Get histories of a certificate file
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param type query string true "Type (options: 'global', 'local')"
// @Param category query string false "Category (directory under 'global' or 'local')"
// @Param name query string true "Script file name"
// @Param page query uint true "Page"
// @Param page_size query uint true "Page size"
// @Success 200 {object} model.PageResult
// @Router /scripts/{host}/log [get]
func (s *CertificateMan) GetScriptLog(c *gin.Context) {
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

	logs, err := s.getScriptLog(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}

	helper.SuccessWithData(c, logs)
}

// @Tags Script
// @Summary Get certificate diff
// @Description Get certificate diff compare to specfied version
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param type query string true "Type (options: 'global', 'local')"
// @Param category query string false "Category (directory under 'global' or 'local')"
// @Param name query string true "Script file name"
// @Param commit query string true "Commit hash"
// @Success 200 {string} string
// @Router /scripts/{host}/diff [get]
func (s *CertificateMan) GetScriptDiff(c *gin.Context) {
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

	diff, err := s.getScriptDiff(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}

	helper.SuccessWithData(c, diff)
}

// @Tags Script
// @Summary Execute certificate
// @Description Execute certificate
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param request body model.ExecuteScript true "Script file creation details"
// @Success 200 {object} model.ScriptResult
// @Router /scripts/{host}/run [post]
func (s *CertificateMan) Execute(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	var req model.ExecuteScript
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	result, err := s.execute(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithData(c, result)
}

// @Tags Script
// @Summary Get run log content
// @Description Get content of run log
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Success 200 {object} model.GitFile
// @Router /scripts/{host}/run/log [get]
func (s *CertificateMan) GetScriptRunLog(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	content, err := s.getScriptRunLog(hostID)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}

	helper.SuccessWithData(c, content)
}

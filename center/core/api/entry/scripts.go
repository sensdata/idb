package entry

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/core/plugin"
	"github.com/sensdata/idb/center/core/plugin/shared"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/model"
)

func getScriptManager() (shared.ScriptManager, error) {
	// 获取插件客户端
	plugin, err := plugin.PLUGINSERVER.GetPlugin("scriptmanager")
	if err != nil {
		return nil, err
	}
	// 类型断言为 gRPC client
	client, ok := plugin.Stub.(shared.ScriptManager)
	if !ok {
		return nil, errors.New("invalid plugin client")
	}
	return client, nil
}

// @Tags Script
// @Summary List category
// @Description List category
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param type query string true "Type (options: 'global', 'local')"
// @Param page query uint true "Page"
// @Param page_size query uint true "Page size"
// @Success 200 {object} model.PageResult
// @Router /scripts/{host}/category [get]
func (b *BaseApi) GetScriptCategories(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	scriptType := c.Query("type")
	if scriptType == "" {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid type", err)
		return
	}
	if scriptType != "global" && scriptType != "local" {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid type", err)
		return
	}

	page, err := strconv.ParseInt(c.Query("page"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid page", err)
		return
	}

	pageSize, err := strconv.ParseInt(c.Query("page_size"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid page_size", err)
		return
	}

	req := model.QueryGitFile{
		Type:     scriptType,
		Category: "",
		Page:     int(page),
		PageSize: int(pageSize),
	}

	// 获取插件
	client, err := getScriptManager()
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	// 调用插件方法
	resp, err := client.ListCategories(hostID, req)
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	SuccessWithData(c, resp)
}

// @Tags Script
// @Summary Create category
// @Description Create category
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param request body model.CreateGitCategory true "Category creation details"
// @Success 200
// @Router /scripts/{host}/category [post]
func (b *BaseApi) CreateScriptCategory(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	var req model.CreateGitCategory
	if err := CheckBindAndValidate(&req, c); err != nil {
		return
	}

	// 获取插件
	client, err := getScriptManager()
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	err = client.CreateCategory(hostID, req)
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}
	SuccessWithData(c, nil)
}

// @Tags Script
// @Summary Update category
// @Description Update category
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param request body model.UpdateGitCategory true "category edit details"
// @Success 200
// @Router /scripts/{host}/category [put]
func (b *BaseApi) UpdateScriptCategory(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	var req model.UpdateGitCategory
	if err := CheckBindAndValidate(&req, c); err != nil {
		return
	}

	// 获取插件
	client, err := getScriptManager()
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	err = client.UpdateCategory(hostID, req)
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}
	SuccessWithData(c, nil)
}

// @Tags Script
// @Summary Delete category
// @Description Delete category
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param type query string true "Type (options: 'global', 'local')"
// @Param category query string true "Category (directory under 'global' or 'local')"
// @Success 200
// @Router /scripts/{host}/category [delete]
func (b *BaseApi) DeleteScriptCategory(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	scriptType := c.Query("type")
	if scriptType == "" {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid type", err)
		return
	}
	if scriptType != "global" && scriptType != "local" {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid type", err)
		return
	}

	category := c.Query("category")
	if category == "" {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid category", err)
		return
	}

	req := model.DeleteGitCategory{
		Type:     scriptType,
		Category: category,
	}

	// 获取插件
	client, err := getScriptManager()
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	err = client.DeleteCategory(hostID, req)
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}
	SuccessWithData(c, nil)
}

// @Tags Script
// @Summary List scripts
// @Description Get list of scripts in a directory
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param type query string true "Type (options: 'global', 'local')"
// @Param category query string true "Category (directory under 'global' or 'local')"
// @Param page query uint true "Page"
// @Param page_size query uint true "Page size"
// @Success 200 {object} model.PageResult
// @Router /scripts/{host} [get]
func (b *BaseApi) GetScriptList(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	scriptType := c.Query("type")
	if scriptType == "" {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid type", err)
		return
	}
	if scriptType != "global" && scriptType != "local" {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid type", err)
		return
	}

	category := c.Query("category")
	if category == "" {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid category", err)
		return
	}

	page, err := strconv.ParseInt(c.Query("page"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid page", err)
		return
	}

	pageSize, err := strconv.ParseInt(c.Query("page_size"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid page_size", err)
		return
	}

	req := model.QueryGitFile{
		Type:     scriptType,
		Category: category,
		Page:     int(page),
		PageSize: int(pageSize),
	}

	// 获取插件
	client, err := getScriptManager()
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	// 调用插件方法
	resp, err := client.ListScripts(hostID, req)
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	SuccessWithData(c, resp)
}

// @Tags Script
// @Summary Get script detail
// @Description Get detail of a script file
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param type query string true "Type (options: 'global', 'local')"
// @Param category query string true "Category (directory under 'global' or 'local')"
// @Param name query string true "Script file name"
// @Success 200 {object} model.GitFile
// @Router /scripts/{host}/detail [get]
func (b *BaseApi) GetScriptDetail(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	scriptType := c.Query("type")
	if scriptType == "" {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid type", err)
		return
	}
	if scriptType != "global" && scriptType != "local" {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid type", err)
		return
	}

	category := c.Query("category")
	if category == "" {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid category", err)
		return
	}

	name := c.Query("name")
	if name == "" {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid name", err)
		return
	}

	req := model.GetGitFileDetail{
		Type:     scriptType,
		Category: category,
		Name:     name,
	}

	// 获取插件
	client, err := getScriptManager()
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	detail, err := client.GetScriptDetail(hostID, req)
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	SuccessWithData(c, detail)
}

// @Tags Script
// @Summary Create script file or category
// @Description Create a new script file or category
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param request body model.CreateGitFile true "Script file creation details"
// @Success 200
// @Router /scripts/{host} [post]
func (b *BaseApi) CreateScript(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	var req model.CreateGitFile
	if err := CheckBindAndValidate(&req, c); err != nil {
		return
	}

	// 获取插件
	client, err := getScriptManager()
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	err = client.CreateScript(hostID, req)
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}
	SuccessWithData(c, nil)
}

// @Tags Script
// @Summary Update script file content
// @Description Update the content of a script file
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param request body model.UpdateGitFile true "Script file edit details"
// @Success 200
// @Router /scripts/{host} [put]
func (b *BaseApi) UpdateScript(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	var req model.UpdateGitFile
	if err := CheckBindAndValidate(&req, c); err != nil {
		return
	}

	// 获取插件
	client, err := getScriptManager()
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	err = client.UpdateScript(hostID, req)
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}
	SuccessWithData(c, nil)
}

// @Tags Script
// @Summary Delete script file
// @Description Delete  a script file
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param type query string true "Type (options: 'global', 'local')"
// @Param category query string true "Category (directory under 'global' or 'local')"
// @Param name query string true "File name"
// @Success 200
// @Router /scripts/{host} [delete]
func (b *BaseApi) DeleteScript(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	scriptType := c.Query("type")
	if scriptType == "" {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid type", err)
		return
	}
	if scriptType != "global" && scriptType != "local" {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid type", err)
		return
	}

	category := c.Query("category")
	if category == "" {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid category", err)
		return
	}

	name := c.Query("name")
	if name == "" {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid name", err)
		return
	}

	req := model.DeleteGitFile{
		Type:     scriptType,
		Category: category,
		Name:     name,
	}

	// 获取插件
	client, err := getScriptManager()
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	err = client.DeleteScript(hostID, req)
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}
	SuccessWithData(c, nil)
}

// @Tags Script
// @Summary Restore script file
// @Description Restore script file to specified version
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param request body model.RestoreGitFile true "Script file restore details"
// @Success 200
// @Router /scripts/{host}/restore [put]
func (b *BaseApi) RestoreScript(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	var req model.RestoreGitFile
	if err := CheckBindAndValidate(&req, c); err != nil {
		return
	}

	// 获取插件
	client, err := getScriptManager()
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	err = client.RestoreScript(hostID, req)
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}
	SuccessWithData(c, nil)
}

// @Tags Script
// @Summary Get script histories
// @Description Get histories of a script file
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param type query string true "Type (options: 'global', 'local')"
// @Param category query string true "Category (directory under 'global' or 'local')"
// @Param name query string true "Script file name"
// @Param page query uint true "Page"
// @Param page_size query uint true "Page size"
// @Success 200 {object} model.PageResult
// @Router /scripts/{host}/history [get]
func (b *BaseApi) GetScriptHistories(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	scriptType := c.Query("type")
	if scriptType == "" {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid type", err)
		return
	}
	if scriptType != "global" && scriptType != "local" {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid type", err)
		return
	}

	category := c.Query("category")
	if category == "" {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid category", err)
		return
	}

	name := c.Query("name")
	if name == "" {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid name", err)
		return
	}

	page, err := strconv.ParseInt(c.Query("page"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid page", err)
		return
	}

	pageSize, err := strconv.ParseInt(c.Query("page_size"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid page_size", err)
		return
	}

	req := model.GitFileLog{
		Type:     scriptType,
		Category: category,
		Name:     name,
		Page:     int(page),
		PageSize: int(pageSize),
	}

	// 获取插件
	client, err := getScriptManager()
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	detail, err := client.GetScriptHistory(hostID, req)
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	SuccessWithData(c, detail)
}

// @Tags Script
// @Summary Get script diff
// @Description Get script diff compare to specfied version
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param type query string true "Type (options: 'global', 'local')"
// @Param category query string true "Category (directory under 'global' or 'local')"
// @Param name query string true "Script file name"
// @Param commit query string true "Commit hash"
// @Success 200 {string} string
// @Router /scripts/{host}/diff [get]
func (b *BaseApi) GetScriptDiff(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	scriptType := c.Query("type")
	if scriptType == "" {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid type", err)
		return
	}
	if scriptType != "global" && scriptType != "local" {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid type", err)
		return
	}

	category := c.Query("category")
	if category == "" {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid category", err)
		return
	}

	name := c.Query("name")
	if name == "" {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid name", err)
		return
	}

	commitHash := c.Query("commit")
	if commitHash == "" {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid commit hash", err)
		return
	}

	req := model.GitFileDiff{
		Type:       scriptType,
		Category:   category,
		Name:       name,
		CommitHash: commitHash,
	}

	// 获取插件
	client, err := getScriptManager()
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	detail, err := client.GetScriptDiff(hostID, req)
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	SuccessWithData(c, detail)
}

// @Tags Script
// @Summary Sync global repository to specified host
// @Description Sync global repository to specified host
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Success 200
// @Router /scripts/{host}/sync [post]
func (b *BaseApi) ScriptSync(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	// 获取插件
	client, err := getScriptManager()
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	err = client.ScriptSync(uint(hostID))
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}
	SuccessWithData(c, nil)
}

// @Tags Script
// @Summary Execute script
// @Description Execute script
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param request body model.ExecuteScript true "Script file creation details"
// @Success 200 {object} model.ScriptResult
// @Router /scripts/{host}/run [post]
func (b *BaseApi) ScriptExec(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	var req model.ExecuteScript
	if err := CheckBindAndValidate(&req, c); err != nil {
		return
	}

	// 获取插件
	client, err := getScriptManager()
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	result, err := client.ScriptExec(uint(hostID), req)
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}
	SuccessWithData(c, result)
}

// @Tags Script
// @Summary Get run logs of script
// @Description Get run logs of script
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param path query string false "Script path"
// @Param page query uint true "Page"
// @Param page_size query uint true "Page size"
// @Success 200 {object} model.PageResult
// @Router /scripts/{host}/run/logs [get]
func (b *BaseApi) GetScriptRunLogs(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	path := c.Query("path")
	if path == "" {
		path = "/"
	}

	page, err := strconv.ParseInt(c.Query("page"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid page", err)
		return
	}

	pageSize, err := strconv.ParseInt(c.Query("page_size"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid page_size", err)
		return
	}

	// 获取插件
	client, err := getScriptManager()
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	result, err := client.GetScriptRunLogs(uint(hostID), path, int(page), int(pageSize))
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	SuccessWithData(c, result)
}

// @Tags Script
// @Summary Get run log detail
// @Description Get run log detail
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param path query string true "Log path"
// @Success 200 {object} model.GitFile
// @Router /scripts/{host}/run/logs/detail [get]
func (b *BaseApi) GetScriptRunLogDetail(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	path := c.Query("path")
	if path == "" {
		path = "/"
	}

	// 获取插件
	client, err := getScriptManager()
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	result, err := client.GetScriptRunLogDetail(uint(hostID), path)
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	SuccessWithData(c, result)
}

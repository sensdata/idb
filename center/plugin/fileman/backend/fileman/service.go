package fileman

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	_ "embed"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/sensdata/idb/center/core/api"
	"github.com/sensdata/idb/center/core/conn"
	"github.com/sensdata/idb/center/global"
	"gopkg.in/yaml.v2"

	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/files"
	"github.com/sensdata/idb/core/helper"
	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/plugin"
)

type FileMan struct {
	config      plugin.PluginConfig
	restyClient *resty.Client
}

var Plugin = FileMan{}

//go:embed plug.yaml
var plugYAML []byte

func (s *FileMan) Initialize() {
	global.LOG.Info("fileman init begin")

	if err := yaml.Unmarshal(plugYAML, &s.config); err != nil {
		global.LOG.Error("Failed to load fileman yaml: %v", err)
		return
	}

	baseUrl := fmt.Sprintf("http://%s:%d/idb/api", "127.0.0.1", conn.CONFMAN.GetConfig().Port)
	global.LOG.Info("baseurl: %s", baseUrl)

	s.restyClient = resty.New().
		SetBaseURL(baseUrl).
		SetHeader("Content-Type", "application/json")

	api.API.SetUpPluginRouters(
		"files",
		[]plugin.PluginRoute{
			{Method: "GET", Path: "/info", Handler: s.GetPluginInfo},
			{Method: "GET", Path: "/menu", Handler: s.GetMenu},
			{Method: "GET", Path: "/trees", Handler: s.GetFileTree},
			{Method: "GET", Path: "", Handler: s.GetFileList},
			{Method: "POST", Path: "", Handler: s.CreateFile},
			{Method: "DELETE", Path: "/:path", Handler: s.DeleteFile},
			{Method: "DELETE", Path: "/batch", Handler: s.BatchDeleteFile},
			{Method: "POST", Path: "/compress", Handler: s.CompressFile},
			{Method: "POST", Path: "/decompress", Handler: s.DeCompressFile},
			{Method: "GET", Path: "/content", Handler: s.GetContent},
			{Method: "PUT", Path: "/content/:path", Handler: s.SaveContent},
			{Method: "POST", Path: "/upload/:host_id/:path", Handler: s.Upload},
			{Method: "GET", Path: "/download/:host_id/:file_path", Handler: s.Download},
			{Method: "GET", Path: "/size", Handler: s.Size},
			{Method: "PUT", Path: "/rename/:host_id/:old_path", Handler: s.ChangeFileName},
			{Method: "PUT", Path: "/move", Handler: s.MoveFile},
			{Method: "PUT", Path: "/owner", Handler: s.ChangeFileOwner},
			{Method: "PUT", Path: "/mode", Handler: s.ChangeFileMode},
			{Method: "PUT", Path: "/batch/role", Handler: s.BatchChangeModeAndOwner},
			{Method: "GET", Path: "/favorites", Handler: s.GetFavoriteList},
			{Method: "POST", Path: "/favorites", Handler: s.CreateFavorite},
			{Method: "DELETE", Path: "/favorites/:id", Handler: s.DeleteFavorite},
		},
	)

	global.LOG.Info("fileman init end")
}

func (s *FileMan) Release() {

}

// @Tags File
// @Summary Get plugin info
// @Description Get plugin information
// @Accept json
// @Produce json
// @Success 200 {object} plugin.PluginInfo
// @Router /files/info [get]
func (s *FileMan) GetPluginInfo(c *gin.Context) {
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
func (s *FileMan) GetMenu(c *gin.Context) {
	menuItems, err := s.getMenus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"type": "menu", "payload": menuItems})
}

func (s *FileMan) getPluginInfo() (plugin.PluginInfo, error) {
	return s.config.Plugin, nil
}

func (s *FileMan) getMenus() ([]plugin.MenuItem, error) {
	return s.config.Menu, nil
}

// @Tags File
// @Summary Get file tree
// @Description Get file tree structure
// @Accept json
// @Produce json
// @Param host_id query uint true "Host ID"
// @Param path query string false "Directory path (default is root directory)"
// @Param page query uint true "Page"
// @Param page_size query uint true "Page size"
// @Success 200 {array} model.FileTree
// @Router /files/trees [get]
func (s *FileMan) GetFileTree(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Query("host_id"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host_id", err)
		return
	}

	path := c.Query("path")
	if path == "" {
		path = "/"
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

	req := model.FileOption{
		HostID: uint(hostID),
		FileOption: files.FileOption{
			Path:     path,
			Expand:   true,
			Page:     int(page),
			PageSize: int(pageSize),
		},
	}

	tree, err := s.getFileTree(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}

	helper.SuccessWithData(c, tree)
}

// @Tags File
// @Summary List files
// @Description Get list of files in a directory
// @Accept json
// @Produce json
// @Param host_id query uint true "Host ID"
// @Param path query string false "Directory path (default is root directory)"
// @Param page query uint true "Page"
// @Param page_size query uint true "Page size"
// @Success 200 {array} model.FileInfo
// @Router /files [get]
func (s *FileMan) GetFileList(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Query("host_id"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host_id", err)
		return
	}

	path := c.Query("path")
	if path == "" {
		path = "/"
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

	req := model.FileOption{
		HostID: uint(hostID),
		FileOption: files.FileOption{
			Path:     path,
			Expand:   true,
			Page:     int(page),
			PageSize: int(pageSize),
		},
	}

	files, err := s.getFileList(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}

	helper.SuccessWithData(c, files)
}

// @Tags File
// @Summary Create file or directory
// @Description Create a new file or directory
// @Accept json
// @Produce json
// @Param request body model.FileCreate true "File creation details"
// @Success 200 {object} model.FileInfo
// @Router /files [post]
func (s *FileMan) CreateFile(c *gin.Context) {
	var req model.FileCreate
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

// @Tags File
// @Summary Delete file
// @Description Delete a file or directory
// @Accept json
// @Produce json
// @Param host_id query uint true "Host ID"
// @Param path path string true "File path"
// @Param force_delete query bool false "Force delete flag"
// @Param is_dir query bool false "Is directory flag"
// @Success 200 "No Content"
// @Router /files/{path} [delete]
func (s *FileMan) DeleteFile(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Query("host_id"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host_id", err)
		return
	}

	path := c.Param("path")
	if path == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Path is required", nil)
		return
	}

	forceDelete, _ := strconv.ParseBool(c.Query("force_delete"))
	isDir, _ := strconv.ParseBool(c.Query("is_dir"))

	req := model.FileDelete{
		HostID:      uint(hostID),
		Path:        path,
		ForceDelete: forceDelete,
		IsDir:       isDir,
	}

	err = s.delete(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}

	c.Status(http.StatusNoContent)
}

// @Tags File
// @Summary Batch delete files
// @Description Delete multiple files or directories
// @Accept json
// @Produce json
// @Param host_id query uint true "Host ID"
// @Param paths query string true "Comma-separated list of file paths to delete"
// @Param is_dir query bool false "Is directory flag"
// @Success 200 "No Content"
// @Router /files/batch [delete]
func (s *FileMan) BatchDeleteFile(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Query("host_id"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host_id", err)
		return
	}

	paths := c.Query("paths")
	if paths == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "No paths provided", nil)
		return
	}

	isDir, _ := strconv.ParseBool(c.Query("is_dir"))

	req := model.FileBatchDelete{
		HostID: uint(hostID),
		Paths:  strings.Split(paths, ","),
		IsDir:  isDir,
	}

	err = s.batchDelete(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}

	c.Status(http.StatusNoContent)
}

// @Tags File
// @Summary Compress files
// @Description Compress multiple files or directories into an archive
// @Accept json
// @Produce json
// @Param request body model.FileCompress true "Compression details"
// @Success 200
// @Router /files/compress [post]
func (s *FileMan) CompressFile(c *gin.Context) {
	var req model.FileCompress
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	err := s.compress(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags File
// @Summary Decompress file
// @Description Decompress file into an archive
// @Accept json
// @Param request body model.FileDeCompress true "request"
// @Success 200
// @Router /files/decompress [post]
func (s *FileMan) DeCompressFile(c *gin.Context) {
	var req model.FileDeCompress
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	err := s.decompress(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags File
// @Summary Get file content
// @Description Get the content of a file
// @Accept json
// @Produce json
// @Param host_id query uint true "Host ID"
// @Param path query string true "File path"
// @Success 200 {object} model.FileInfo
// @Router /files/content [get]
func (s *FileMan) GetContent(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Query("host_id"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host_id", err)
		return
	}

	path := c.Query("path")
	if path == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Path is required", nil)
		return
	}

	req := model.FileContentReq{
		HostID: uint(hostID),
		Path:   path,
	}

	info, err := s.getContent(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithData(c, info)
}

// @Tags File
// @Summary Update file content
// @Description Update the content of a file
// @Accept json
// @Produce json
// @Param host_id query uint true "Host ID"
// @Param path path string true "File path"
// @Param content body string true "New file content"
// @Success 200
// @Router /files/content/{path} [put]
func (s *FileMan) SaveContent(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Query("host_id"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host_id", err)
		return
	}

	path := c.Param("path")
	if path == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Path is required", nil)
		return
	}

	var content string
	if err := c.ShouldBindJSON(&content); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid content", err)
		return
	}

	req := model.FileEdit{
		HostID:  uint(hostID),
		Path:    path,
		Content: content,
	}

	if err := s.saveContent(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags File
// @Summary Upload file
// @Description Upload a file to a specific host and path
// @Accept multipart/form-data
// @Produce json
// @Param host_id path int true "Host ID"
// @Param path path string true "Destination directory path"
// @Param file formData file true "File to upload"
// @Success 200 "No Content"
// @Router /files/upload/{host_id}/{path} [post]
func (s *FileMan) Upload(c *gin.Context) {
	hostID, err := strconv.Atoi(c.Param("host_id"))
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host_id", err)
		return
	}

	path := c.Param("path")
	if path == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Path is required", nil)
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid file", err)
		return
	}

	if err := s.uploadFile(uint(hostID), path, file); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags File
// @Summary Download file
// @Description Download a file from a specific host and path
// @Produce octet-stream
// @Param host_id path int true "Host ID"
// @Param file_path path string true "File path"
// @Success 200 {file} binary
// @Router /files/download/{host_id}/{file_path} [get]
func (s *FileMan) Download(c *gin.Context) {
	hostID, err := strconv.Atoi(c.Param("host_id"))
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host_id", err)
		return
	}

	filePath := c.Param("file_path")
	if filePath == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "File path is required", nil)
		return
	}

	if err := s.downloadFile(c, uint(hostID), filePath); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
}

// @Tags File
// @Summary Get directory size
// @Description Get the size of a directory
// @Accept json
// @Produce json
// @Param host_id query uint true "Host ID"
// @Param path query string true "Directory path"
// @Success 200 {object} model.DirSizeRes
// @Router /files/size [get]
func (s *FileMan) Size(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Query("host_id"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host_id", err)
		return
	}

	path := c.Query("path")
	if path == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Path is required", nil)
		return
	}

	req := model.DirSizeReq{
		HostID: uint(hostID),
		Path:   path,
	}

	res, err := s.dirSize(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithData(c, res)
}

// @Tags File
// @Summary Change file name
// @Description Rename a file
// @Accept json
// @Produce json
// @Param host_id path int true "Host ID"
// @Param old_path path string true "Current file path"
// @Param new_name body string true "New file name"
// @Success 200 "No Content"
// @Router /files/rename/{host_id}/{old_path} [put]
func (s *FileMan) ChangeFileName(c *gin.Context) {
	hostID, err := strconv.Atoi(c.Param("host_id"))
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host_id", err)
		return
	}

	oldPath := c.Param("old_path")
	if oldPath == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Old path is required", nil)
		return
	}

	var newName string
	if err := c.ShouldBindJSON(&newName); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid new name", err)
		return
	}

	req := model.FileRename{
		HostID:  uint(hostID),
		Path:    oldPath,
		NewName: newName,
	}

	if err := s.changeName(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// // @Tags File
// // @Summary Wget file
// // @Description 下载远端文件
// // @Accept json
// // @Param request body model.FileWget true "request"
// // @Success 200
// // @Router /files/wget [post]
// func (s *FileMan) WgetFile(c *gin.Context) {
// 	var req model.FileWget
// 	if err := helper.CheckBindAndValidate(&req, c); err != nil {
// 		return
// 	}
// 	helper.ErrorWithDetail(c, constant.CodeFailed, "not supported yet", nil)
// }

// @Tags File
// @Summary Move files
// @Description Move one or more files to a new location
// @Accept json
// @Produce json
// @Param move_request body model.FileMove true "File move details"
// @Success 200 "No Content"
// @Router /files/move [put]
func (s *FileMan) MoveFile(c *gin.Context) {
	var req model.FileMove
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := s.mvFile(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags File
// @Summary Change file owner
// @Description Change file user or/and group
// @Accept json
// @Produce json
// @Param request body model.FileRoleUpdate true "File owner update details"
// @Success 200 "No Content"
// @Router /files/owner [put]
func (s *FileMan) ChangeFileOwner(c *gin.Context) {
	var req model.FileRoleUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := s.changeOwner(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags File
// @Summary Change file mode
// @Description Change file mode
// @Accept json
// @Produce json
// @Param request body model.FileCreate true "File mode update details"
// @Success 200 "No Content"
// @Router /files/mode [put]
func (s *FileMan) ChangeFileMode(c *gin.Context) {
	var req model.FileCreate
	if err := helper.CheckBind(&req, c); err != nil {
		return
	}
	err := s.changeMode(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags File
// @Summary Batch change file mode and owner
// @Description Batch modify file permissions and user/group
// @Accept json
// @Produce json
// @Param request body model.FileRoleReq true "Batch file mode and owner change request"
// @Success 200 "No Content"
// @Router /files/batch/role [put]
func (s *FileMan) BatchChangeModeAndOwner(c *gin.Context) {
	var req model.FileRoleReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := s.batchChangeModeAndOwner(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
	}
	helper.SuccessWithOutData(c)
}

// @Tags File
// @Summary Get favorites
// @Description Get favorite files
// @Accept json
// @Produce json
// @Param request body model.FavoriteListReq true "Favorite list request"
// @Success 200 {object} model.PageResult
// @Router /files/favorites [get]
func (s *FileMan) GetFavoriteList(c *gin.Context) {
	var req model.FavoriteListReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	result, err := s.getFavoriteList(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithData(c, result)
}

// @Tags File
// @Summary Collect a favorite file
// @Description Collect a favorite file
// @Accept json
// @Produce json
// @Param request body model.FavoriteCreate true "Favorite create request"
// @Success 200 {object} model.Favorite
// @Router /files/favorites [post]
func (s *FileMan) CreateFavorite(c *gin.Context) {
	var req model.FavoriteCreate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	favorite, err := s.createFavorite(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithData(c, favorite)
}

// @Tags File
// @Summary Delete a favorite
// @Description Delete a favorite
// @Accept json
// @Produce json
// @Param host_id query uint true "Host ID"
// @Param id path uint true "Favorite ID"
// @Success 200 "No Content"
// @Router /files/favorites/{id} [delete]
func (s *FileMan) DeleteFavorite(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Query("host_id"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host_id", err)
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid favorite id", err)
		return
	}

	req := model.FavoriteDelete{
		HostID: uint(hostID),
		ID:     uint(id),
	}

	if err := s.deleteFavorite(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithOutData(c)
}

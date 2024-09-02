package fileman

import (
	"net/http"

	_ "embed"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/sensdata/idb/center/core/api"
	"github.com/sensdata/idb/center/global"
	"gopkg.in/yaml.v2"

	"github.com/sensdata/idb/core/constant"
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

	s.restyClient = resty.New().
		SetBaseURL("http://127.0.0.1:8080").
		SetHeader("Content-Type", "application/json")

	api.API.SetUpPluginRouters(
		"files",
		[]plugin.PluginRoute{
			{Method: "GET", Path: "/info", Handler: s.GetPluginInfo},
			{Method: "GET", Path: "/menu", Handler: s.GetMenu},
			{Method: "POST", Path: "/tree", Handler: s.GetFileTree},
			{Method: "POST", Path: "/list", Handler: s.GetFileList},
			{Method: "POST", Path: "/create", Handler: s.CreateFile},
			{Method: "POST", Path: "/del", Handler: s.DeleteFile},
			{Method: "POST", Path: "/batch/del", Handler: s.BatchDeleteFile},
			{Method: "POST", Path: "/compress", Handler: s.CompressFile},
			{Method: "POST", Path: "/decompress", Handler: s.DeCompressFile},
			{Method: "POST", Path: "/content", Handler: s.GetContent},
			{Method: "POST", Path: "/content/save", Handler: s.SaveContent},
			{Method: "POST", Path: "/upload", Handler: s.UploadFiles},
			{Method: "POST", Path: "/download", Handler: s.Download},
			{Method: "POST", Path: "/wget", Handler: s.WgetFile},
			{Method: "POST", Path: "/size", Handler: s.Size},
			{Method: "POST", Path: "/rename", Handler: s.ChangeFileName},
			{Method: "POST", Path: "/move", Handler: s.MoveFile},
			{Method: "POST", Path: "/owner", Handler: s.ChangeFileOwner},
			{Method: "POST", Path: "/mode", Handler: s.ChangeFileMode},
			{Method: "POST", Path: "/batch/role", Handler: s.BatchChangeModeAndOwner},
			{Method: "POST", Path: "/favorite/list", Handler: s.GetFavoriteList},
			{Method: "POST", Path: "/favorite/create", Handler: s.CreateFavorite},
			{Method: "POST", Path: "/favorite/del", Handler: s.DeleteFavorite},
		},
	)

	global.LOG.Info("fileman init end")
}

func (s *FileMan) Release() {

}

// GetPluginInfo 处理 /fileman/plugin_info 请求
func (s *FileMan) GetPluginInfo(c *gin.Context) {
	pluginInfo, err := s.getPluginInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"type": "info", "payload": pluginInfo})
}

// GetMenu 处理 /fileman/menu 请求
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
// @Summary Load files tree
// @Description 加载文件树
// @Accept json
// @Param request body request.FileOption true "request"
// @Success 200 {array} response.FileTree
// @Router /files/tree [post]
func (s *FileMan) GetFileTree(c *gin.Context) {
	var req model.FileOption
	if err := helper.CheckBind(&req, c); err != nil {
		return
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
// @Description 获取文件列表
// @Accept json
// @Param request body request.FileOption true "request"
// @Success 200 {object} response.FileInfo
// @Router /files/list [post]
func (s *FileMan) GetFileList(c *gin.Context) {
	var req model.FileOption
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	files, err := s.getFileList(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithData(c, files)
}

// @Tags File
// @Summary Create file
// @Description 创建文件/文件夹
// @Accept json
// @Param request body request.FileCreate true "request"
// @Success 200
// @Router /files/create [post]
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
// @Description 删除文件/文件夹
// @Accept json
// @Param request body request.FileDelete true "request"
// @Success 200
// @Router /files/del [post]
// @x-panel-log {"bodyKeys":["path"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"删除文件/文件夹 [path]","formatEN":"Delete dir or file [path]"}
func (s *FileMan) DeleteFile(c *gin.Context) {
	var req model.FileDelete
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	err := s.delete(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags File
// @Summary Batch delete file
// @Description 批量删除文件/文件夹
// @Accept json
// @Param request body request.FileBatchDelete true "request"
// @Success 200
// @Router /files/batch/del [post]
func (s *FileMan) BatchDeleteFile(c *gin.Context) {
	var req model.FileBatchDelete
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	err := s.batchDelete(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags File
// @Summary Compress file
// @Description 压缩文件
// @Accept json
// @Param request body request.FileCompress true "request"
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
// @Description 解压文件
// @Accept json
// @Param request body request.FileDeCompress true "request"
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
// @Summary Load file content
// @Description 获取文件内容
// @Accept json
// @Param request body request.FileContentReq true "request"
// @Success 200 {object} response.FileInfo
// @Router /files/content [post]
func (s *FileMan) GetContent(c *gin.Context) {
	var req model.FileContentReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
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
// @Description 更新文件内容
// @Accept json
// @Param request body request.FileEdit true "request"
// @Success 200
// @Router /files/save [post]
func (s *FileMan) SaveContent(c *gin.Context) {
	var req model.FileEdit
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := s.saveContent(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags File
// @Summary Upload file
// @Description 上传文件
// @Param file formData file true "request"
// @Success 200
// @Router /files/upload [post]
func (s *FileMan) UploadFiles(c *gin.Context) {
	helper.ErrorWithDetail(c, constant.CodeFailed, "not supported yet", nil)
}

// @Tags File
// @Summary Download file
// @Description 下载文件
// @Accept json
// @Success 200
// @Router /files/download [get]
func (s *FileMan) Download(c *gin.Context) {
	helper.ErrorWithDetail(c, constant.CodeFailed, "not supported yet", nil)
}

// @Tags File
// @Summary Load file size
// @Description 获取文件夹大小
// @Accept json
// @Param request body request.DirSizeReq true "request"
// @Success 200
// @Router /files/size [post]
func (s *FileMan) Size(c *gin.Context) {
	var req model.DirSizeReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
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
// @Description 修改文件名称
// @Accept json
// @Param request body request.FileRename true "request"
// @Success 200
// @Router /files/rename [post]
func (s *FileMan) ChangeFileName(c *gin.Context) {
	var req model.FileRename
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := s.changeName(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags File
// @Summary Wget file
// @Description 下载远端文件
// @Accept json
// @Param request body request.FileWget true "request"
// @Success 200
// @Router /files/wget [post]
func (s *FileMan) WgetFile(c *gin.Context) {
	var req model.FileWget
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	helper.ErrorWithDetail(c, constant.CodeFailed, "not supported yet", nil)
}

// @Tags File
// @Summary Move file
// @Description 移动文件
// @Accept json
// @Param request body request.FileMove true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /files/move [post]
// @x-panel-log {"bodyKeys":["oldPaths","newPath"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"移动文件 [oldPaths] => [newPath]","formatEN":"Move [oldPaths] => [newPath]"}
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
// @Description 修改文件用户/组
// @Accept json
// @Param request body request.FileRoleUpdate true "request"
// @Success 200
// @Router /files/owner [post]
// @x-panel-log {"bodyKeys":["path","user","group"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"修改用户/组 [paths] => [user]/[group]","formatEN":"Change owner [paths] => [user]/[group]"}
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
// @Description 修改文件权限
// @Accept json
// @Param request body request.FileCreate true "request"
// @Success 200
// @Router /files/mode [post]
func (s *FileMan) ChangeFileMode(c *gin.Context) {
	var req model.FileCreate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
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
// @Description 批量修改文件权限和用户/组
// @Accept json
// @Param request body request.FileRoleReq true "request"
// @Success 200
// @Router /files/batch/role [post]
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
// @Summary List favorites
// @Description 获取收藏列表
// @Accept json
// @Param request body dto.PageInfo true "request"
// @Success 200
// @Router /files/favorite/list [post]
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
// @Summary Create favorite
// @Description 创建收藏
// @Accept json
// @Param request body request.FavoriteCreate true "request"
// @Success 200
// @Router /files/favorite/create [post]
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
// @Summary Delete favorite
// @Description 删除收藏
// @Accept json
// @Param request body request.FavoriteDelete true "request"
// @Router /files/favorite/del [post]
func (s *FileMan) DeleteFavorite(c *gin.Context) {
	var req model.FavoriteDelete
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := s.deleteFavorite(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	helper.SuccessWithOutData(c)
}

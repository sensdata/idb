package entry

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/model"
)

// @Tags App
// @Summary Sync app list
// @Description 同步应用列表
// @Success 200
// @Router /store/apps/sync [post]
func (b *BaseApi) SyncApp(c *gin.Context) {
	err := appService.SyncApp()
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	SuccessWithData(c, "")
}

// @Tags App
// @Summary Get app list
// @Description Get app list
// @Param page query int true "Page number"
// @Param page_size query int true "Page size"
// @Param name query string false "Name"
// @Param category query string false "Category"
// @Success 200
// @Router /store/apps [get]
func (b *BaseApi) AppPage(c *gin.Context) {
	var req model.QueryApp
	if err := CheckQueryAndValidate(&req, c); err != nil {
		return
	}

	result, err := appService.AppPage(req)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	SuccessWithData(c, result)
}

// @Tags App
// @Summary Get app detail
// @Description Get app detail
// @Param id path int true "App ID"
// @Success 200
// @Router /store/apps/:id [get]
func (b *BaseApi) AppDetail(c *gin.Context) {
	appID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid app ID", err)
		return
	}

	result, err := appService.AppDetail(model.QueryAppDetail{ID: uint(appID)})
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	SuccessWithData(c, result)
}

// @Tags App
// @Summary Get installed app list
// @Description Get installed app list
// @Param host path int true "Host ID"
// @Param page query int true "Page number"
// @Param page_size query int true "Page size"
// @Param name query string false "Name"
// @Success 200
// @Router /store/:host/apps/installed [get]
func (b *BaseApi) InstalledAppPage(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	var req model.QueryInstalledApp
	if err := CheckQueryAndValidate(&req, c); err != nil {
		return
	}

	result, err := appService.InstalledAppPage(hostID, req)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	SuccessWithData(c, result)
}

// @Tags App
// @Summary Install app
// @Description Install app
// @Param host path int true "Host ID"
// @Param request body model.ComposeCreate true "request"
// @Success 200 {object} model.ComposeCreateResult
// @Router /store/:host/apps/install [post]
func (b *BaseApi) InstallApp(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	var req model.ComposeCreate
	if err := CheckBindAndValidate(&req, c); err != nil {
		return
	}

	result, err := appService.AppInstall(uint64(hostID), req)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	SuccessWithData(c, result)
}

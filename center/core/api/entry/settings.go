package entry

import (
	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/model"
)

// @Tags Settings
// @Summary Get server descriptions
// @Description Get server descriptions
// @Accept json
// @Produce json
// @Success 200 {object} model.About
// @Router /settings/about [get]
func (b *BaseApi) About(c *gin.Context) {
	result, err := settingsService.About()
	if err != nil {
		ErrorWithDetail(c, constant.CodeSuccess, constant.ErrNoRecords.Error(), err)
		return
	}
	SuccessWithData(c, result)
}

// @Tags Settings
// @Summary Get avaiable ips
// @Description Get avaiable ips
// @Accept json
// @Produce json
// @Success 200 {object} model.AvailableIps
// @Router /settings/ips [get]
func (b *BaseApi) IPs(c *gin.Context) {
	result, err := settingsService.IPs()
	if err != nil {
		ErrorWithDetail(c, constant.CodeSuccess, constant.ErrNoRecords.Error(), err)
		return
	}
	SuccessWithData(c, result)
}

// @Tags Settings
// @Summary Get avaiable timezones
// @Description Get avaiable timezones
// @Accept json
// @Produce json
// @Param page query int true "Page number"
// @Param page_size query int true "Page size"
// @Success 200 {object} model.PageResult
// @Router /settings/timezones [get]
func (b *BaseApi) Timezones(c *gin.Context) {

	var req model.SearchPageInfo
	if err := CheckQueryAndValidate(&req, c); err != nil {
		return
	}

	result, err := settingsService.Timezones(req)
	if err != nil {
		ErrorWithDetail(c, constant.CodeSuccess, constant.ErrNoRecords.Error(), err)
		return
	}
	SuccessWithData(c, result)
}

// @Tags Settings
// @Summary Get server settings
// @Description Get server settings
// @Accept json
// @Produce json
// @Success 200 {object} model.SettingInfo
// @Router /settings [get]
func (b *BaseApi) Settings(c *gin.Context) {
	result, err := settingsService.Settings()
	if err != nil {
		ErrorWithDetail(c, constant.CodeSuccess, constant.ErrNoRecords.Error(), err)
		return
	}
	SuccessWithData(c, result)
}

// @Tags Settings
// @Summary Update server settings
// @Description Update server settings
// @Accept json
// @Produce json
// @Param request body model.UpdateSettingRequest true "request"
// @Success 200 {object} model.UpdateSettingResponse
// @Router /settings [post]
func (b *BaseApi) UpdateSettings(c *gin.Context) {
	var req model.UpdateSettingRequest
	if err := CheckBindAndValidate(&req, c); err != nil {
		return
	}

	rsp, err := settingsService.Update(req)
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}
	SuccessWithData(c, rsp)
}

// @Tags Settings
// @Summary Upgrade server
// @Description Upgrade server
// @Accept json
// @Produce json
// @Success 200
// @Router /settings/upgrade [post]
func (b *BaseApi) Upgrade(c *gin.Context) {
	err := settingsService.Upgrade()
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}
	SuccessWithData(c, nil)
}

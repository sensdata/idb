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
// @Success 200
// @Router /settings [post]
func (b *BaseApi) UpdateSettings(c *gin.Context) {
	var req model.UpdateSettingRequest
	if err := CheckBindAndValidate(&req, c); err != nil {
		return
	}

	err := settingsService.Update(req)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	SuccessWithData(c, "")
}

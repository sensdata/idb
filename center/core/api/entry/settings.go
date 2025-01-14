package entry

import (
	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/utils"
)

// @Tags Settings
// @Summary Get profile
// @Description Get profile
// @Accept json
// @Produce json
// @Success 200 {object} model.Profile
// @Router /settings/profile [get]
func (b *BaseApi) Profile(c *gin.Context) {
	claimsInterface, exists := c.Get("user")
	if !exists {
		ErrorWithDetail(c, constant.CodeAuth, constant.ErrAuth.Error(), constant.ErrAuth)
		return
	}
	claims, ok := claimsInterface.(utils.Claims)
	if !ok {
		ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), constant.ErrInternalServer)
		return
	}
	result, err := settingsService.Profile(claims.ID)
	if err != nil {
		ErrorWithDetail(c, constant.CodeSuccess, constant.ErrNoRecords.Error(), err)
		return
	}
	SuccessWithData(c, result)
}

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

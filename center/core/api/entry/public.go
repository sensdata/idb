package entry

import (
	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/core/constant"
)

// @Tags Settings
// @Summary Get server descriptions
// @Description Get server descriptions
// @Accept json
// @Produce json
// @Success 200 {object} model.About
// @Router /public/version [get]
func (b *BaseApi) Version(c *gin.Context) {
	result, err := publicService.Version()
	if err != nil {
		ErrorWithDetail(c, constant.CodeSuccess, constant.ErrNoRecords.Error(), err)
		return
	}
	SuccessWithData(c, result)
}

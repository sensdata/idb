package entry

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/core/constant"
)

// @Tags Home
// @Summary Get managed app list
// @Description Get managed app list
// @Param host path int true "Host ID"
// @Success 200 {object} model.PageResult
// @Router /home/{host}/managed/apps [get]
func (b *BaseApi) ManagedApps(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	result, err := appService.ManagedApps(hostID)
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}
	SuccessWithData(c, result)
}

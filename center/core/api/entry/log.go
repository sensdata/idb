package entry

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/global"
)

// @Tags Log
// @Summary Connect to log stream
// @Description Connect to log stream through Server-Sent Events
// @Accept json
// @Produce text/event-stream
// @Param host path uint true "Host ID"
// @Param path query string true "File path"
// @Param whence query string false "Whence, one of 'start', 'end'"
// @Success 200 {string} string "SSE stream started"
// @Failure 400 {object} model.Response "Bad Request"
// @Router /logs/{host}/follow [get]
func (b *BaseApi) HandleLogStream(c *gin.Context) {
	err := logManService.HandleLogStream(c)
	if err != nil {
		global.LOG.Error("Handle log stream failed: %v", err)
		ErrorWithDetail(c, http.StatusInternalServerError, "Failed to establish SSE connection", err)
		return
	}
}

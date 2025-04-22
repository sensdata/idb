package entry

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/core/conn"
	"github.com/sensdata/idb/center/global"
)

// @Tags Task
// @Summary Connect to task log stream
// @Description Connect to task log stream through Server-Sent Events
// @Accept json
// @Produce text/event-stream
// @Param taskId path string true "Task ID"
// @Param offset query int false "Offset"
// @Param whence query string false "Whence, one of 'start', 'end'"
// @Success 200 {string} string "SSE stream started"
// @Failure 400 {object} model.Response "Bad Request"
// @Router /tasks/{taskId}/logs [get]
func (b *BaseApi) HandleTaskLogStream(c *gin.Context) {
	err := conn.TASK.HandleTaskLogStream(c)
	if err != nil {
		global.LOG.Error("Handle task log stream failed: %v", err)
		ErrorWithDetail(c, http.StatusInternalServerError, "Failed to establish SSE connection", err)
		return
	}
}

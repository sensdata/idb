package entry

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/core/conn"
	"github.com/sensdata/idb/center/global"
)

// @Tags Terminal
// @Summary SSH Terminal
// @Description SSH Terminal
// @Accept json
// @Produce json
// @Param host_id query uint true "Host ID"
// @Param cols query uint false "Window cols, default 80"
// @Param rows query uint false "Window rows, default 40"
// @Success 101 {string} string "Switching Protocols to websocket"
// @Failure 400 {object} model.Response "Bad Request"
// @Router /ws/terminals/ssh [get]
func (b *BaseApi) HandleSshTerminal(c *gin.Context) {
	err := conn.WEBSOCKET.HandleSshTerminal(c)
	if err != nil {
		global.LOG.Error("Handle ssh terminal failed: %v", err)

		// 检查是否为升级错误
		if err.Error() == "websocket: the client is not using the websocket protocol: 'upgrade' token not found in 'Connection' header" {
			ErrorWithDetail(c, http.StatusBadRequest, "Failed to establish WebSocket connection", err)
		} else {
			// 对于其他错误，返回一个不同的状态码
			ErrorWithDetail(c, http.StatusInternalServerError, "Internal server error", err)
		}
		return
	}

	// 如果没有错误，WebSocket连接已经建立，不需要做任何事情
}

// @Tags Terminal
// @Summary Agent Terminal
// @Description Agent Terminal
// @Accept json
// @Produce json
// @Param host_id query uint true "Host ID"
// @Param request body model.TerminalMessage true "request"
// @Success 200
// @Router /ws/terminals/agent [get]
func (b *BaseApi) HandleAgentTerminal(c *gin.Context) {
	err := conn.WEBSOCKET.HandleAgentTerminal(c)
	if err != nil {
		global.LOG.Error("Handle agent terminal failed: %v", err)

		// 检查是否为升级错误
		if err.Error() == "websocket: the client is not using the websocket protocol: 'upgrade' token not found in 'Connection' header" {
			ErrorWithDetail(c, http.StatusBadRequest, "Failed to establish WebSocket connection", err)
		} else {
			// 对于其他错误，返回一个不同的状态码
			ErrorWithDetail(c, http.StatusInternalServerError, "Internal server error", err)
		}
		return
	}

	// 如果没有错误，WebSocket连接已经建立，不需要做任何事情
}

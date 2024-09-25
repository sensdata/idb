package entry

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/core/conn"
	"github.com/sensdata/idb/center/global"
)

// @Tags WebSocket
// @Summary 终端会话接口
// @Description 终端会话的websocket接口
// @Accept json
// @Produce json
// @Success 101 {string} string "Switching Protocols to websocket"
// @Failure 400 {object} model.Response "Bad Request"
// @Router /ws/terminals [get]
func (b *BaseApi) HandleTerminal(c *gin.Context) {
	err := conn.WEBSOCKET.HandleTerminal(c)
	if err != nil {
		global.LOG.Error("Handle terminal failed: " + err.Error())

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

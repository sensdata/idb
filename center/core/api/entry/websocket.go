package entry

import (
	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/core/conn"
	"github.com/sensdata/idb/center/global"
)

// @Tags WebSocket
// @Summary 终端会话接口
// @Description 终端会话的websocket接口
// @Accept query
// @Produce stream
// @Param id hostid string true ""
// @Success 200
// @Router /idb/ws/terminal [get]
func (b *BaseApi) HandleTerminal(c *gin.Context) {
	err := conn.WEBSOCKET.HandleTerminal(c)
	if err != nil {
		global.LOG.Error("Handle terminal failed")
	}
}

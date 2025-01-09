package entry

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/core/conn"
	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/model"
)

// @Tags Terminal(ssh)
// @Summary Connect to ssh terminal
// @Description Connect to ssh terminal
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param cols query uint false "Window cols, default 80"
// @Param rows query uint false "Window rows, default 40"
// @Success 101 {string} string "Switching Protocols to websocket"
// @Failure 400 {object} model.Response "Bad Request"
// @Router /terminals/{host}/ssh/start [get]
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
}

// @Tags Terminal
// @Summary Create or reconnect to a terminal session through websocket
// @Description Create or reconnect to a terminal session through websocket. When starting a session, session name can be specified via model.TerminalMessage.Data, or system will generate one automatically. When attaching to a session, session ID can be specified via model.TerminalMessage.Session, or system will create a new session or attach to a recent one automatically.
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param request body model.TerminalMessage true "request data send to websocket"
// @Success 200
// @Router /terminals/{host}/start [get]
func (b *BaseApi) HandleTerminal(c *gin.Context) {
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
}

// @Tags Terminal
// @Summary Get agent terminal sessions
// @Description Get agent terminal sessions
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Success 200
// @Router /terminals/{host}/sessions [get]
func (b *BaseApi) TerminalSessions(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	result, err := terminalService.Sessions(uint(hostID))
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	SuccessWithData(c, result)
}

// @Tags Terminal
// @Summary Detach terminal session
// @Description Detach terminal session
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param request body model.TerminalRequest true "Request details"
// @Success 200
// @Router /terminals/{host}/session/detach [post]
func (b *BaseApi) DetachSession(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	var req model.TerminalRequest
	if err := CheckQueryAndValidate(&req, c); err != nil {
		return
	}

	err = terminalService.Detach(uint(hostID), req)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	SuccessWithData(c, "")
}

// @Tags Terminal
// @Summary Quit terminal session
// @Description Quit terminal session
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param request body model.TerminalRequest true "Request details"
// @Success 200
// @Router /terminals/{host}/session/quit [post]
func (b *BaseApi) QuitSession(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	var req model.TerminalRequest
	if err := CheckQueryAndValidate(&req, c); err != nil {
		return
	}

	err = terminalService.Quit(uint(hostID), req)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	SuccessWithData(c, "")
}

// @Tags Terminal
// @Summary Rename terminal session
// @Description Rename terminal session
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param request body model.TerminalRequest true "Request details"
// @Success 200
// @Router /terminals/{host}/session/rename [post]
func (b *BaseApi) RenameSession(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	var req model.TerminalRequest
	if err := CheckQueryAndValidate(&req, c); err != nil {
		return
	}

	err = terminalService.Rename(uint(hostID), req)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	SuccessWithData(c, "")
}

// @Tags Terminal
// @Summary Install terminal
// @Description Install terminal
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Success 200
// @Router /terminals/{host}/install [post]
func (b *BaseApi) InstallTerminal(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host id", err)
		return
	}

	err = terminalService.Install(uint(hostID))
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), err)
		return
	}
	SuccessWithData(c, "")
}

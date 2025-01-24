package message

const (
	WsMessageCmd       = "cmd"
	WsMessageResize    = "resize"
	WsMessageHeartbeat = "heartbeat"
	WsMessageStart     = "start"
	WsMessageAttach    = "attach"
)

type WsMessage struct {
	Code      int    `json:"code"`
	Msg       string `json:"msg"`
	Type      string `json:"type"`
	Session   string `json:"session"`
	Data      string `json:"data"`
	Cols      int    `json:"cols"`
	Rows      int    `json:"rows"`
	Timestamp int    `json:"timestamp"`
}

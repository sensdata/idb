package message

const (
	WsMessageCmd       = "cmd"
	WsMessageResize    = "resize"
	WsMessageHeartbeat = "heartbeat"
	WsMessageStart     = "start"
	WsMessageAttach    = "attach"
)

type WsMessage struct {
	Type      string `json:"type"`
	Session   string `json:"session,omitempty"`
	Data      string `json:"data,omitempty"`
	Cols      int    `json:"cols,omitempty"`
	Rows      int    `json:"rows,omitempty"`
	Timestamp int    `json:"timestamp,omitempty"`
}

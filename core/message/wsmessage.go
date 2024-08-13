package message

const (
	WsMessageCmd       = "cmd"
	WsMessageResize    = "resize"
	WsMessageHeartbeat = "heartbeat"
)

type WsMessage struct {
	Type      string `json:"type"`
	Data      string `json:"data,omitempty"`      // WsMessageCmd
	Cols      int    `json:"cols,omitempty"`      // WsMessageResize
	Rows      int    `json:"rows,omitempty"`      // WsMessageResize
	Timestamp int    `json:"timestamp,omitempty"` // WsMessageHeartbeat
}

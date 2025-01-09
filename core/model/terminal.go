package model

import "time"

type TerminalMessage struct {
	Type      string `json:"type" validate:"required,oneof=start attach command"`
	Timestamp int64  `json:"timestamp"`
	Session   string `json:"session,omitempty"`
	Data      string `json:"data,omitempty"`
}

type TerminalRequest struct {
	Session string `json:"session"`
	Data    string `json:"data,omitempty"`
}

type SessionInfo struct {
	Session string    `json:"session"`
	Name    string    `json:"name"`
	Time    time.Time `json:"time"`
	Status  string    `json:"status"`
}

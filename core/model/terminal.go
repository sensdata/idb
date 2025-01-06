package model

import "time"

type TerminalMessage struct {
	Type    string `json:"type" validate:"required,oneof=start attach command"`
	Session string `json:"session"`
	Data    string `json:"data,omitempty"`
}

type TerminalRequest struct {
	Session string `json:"session"`
	Data    string `json:"data,omitempty"`
}

type SessionInfo struct {
	Session string    `json:"session"`
	Time    time.Time `json:"time"`
	Status  string    `json:"status"`
}

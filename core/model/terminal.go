package model

import (
	"time"
)

type TerminalRequest struct {
	Type    string `json:"type" validate:"required,oneof=screen tmux docker"`
	Session string `json:"session"`
	Data    string `json:"data"`
}

type SessionInfo struct {
	Session string    `json:"session"`
	Name    string    `json:"name"`
	Time    time.Time `json:"time"`
	Status  string    `json:"status"`
}

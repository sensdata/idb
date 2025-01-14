package model

import "time"

type TerminalRequest struct {
	Session string `json:"session"`
	Data    string `json:"data"`
}

type SessionInfo struct {
	Session string    `json:"session"`
	Name    string    `json:"name"`
	Time    time.Time `json:"time"`
	Status  string    `json:"status"`
}

package model

type TerminalMessage struct {
	Type    string `json:"type" validate:"required,oneof=start command"`
	Session string `json:"session"`
	Data    string `json:"data,omitempty"`
}

package model

import "time"

type ExecuteScript struct {
	ScriptPath string `json:"script_path" validate:"required"`
}

type ScriptExec struct {
	ScriptPath string `json:"script_path"`
	LogPath    string `json:"log_path"`
}

type ScriptResult struct {
	LogHost uint      `json:"log_host"`
	LogPath string    `json:"log_path"`
	Start   time.Time `json:"start"`
	End     time.Time `json:"end"`
	Out     string    `json:"out"`
	Err     string    `json:"err"`
}

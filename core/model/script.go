package model

import "time"

type ExecuteScript struct {
	HostID     uint   `json:"host_id" validate:"required"`
	ScriptPath string `json:"script_path" validate:"required"`
}

type ScriptExec struct {
	ScriptPath string `json:"script_path"`
	LogPath    string `json:"log_path"`
}

type ScriptResult struct {
	TaskID  string    `json:"task_id"`
	LogPath string    `json:"log_path"`
	Start   time.Time `json:"start"`
	End     time.Time `json:"end"`
	Out     string    `json:"out"`
	Err     string    `json:"err"`
}

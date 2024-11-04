package model

import "time"

type ScriptInfo struct {
	Source     string    `json:"source"`
	Type       string    `json:"type"`
	Category   string    `json:"category"`
	Name       string    `json:"name"`
	Extension  string    `json:"extension"`
	Content    string    `json:"content"`
	Size       int64     `json:"size"`
	UpdateTime time.Time `json:"update_time"`
	ModTime    time.Time `json:"mod_time"`
}

type ExecuteScript struct {
	HostID     uint   `json:"host_id" validate:"required"`
	ScriptPath string `json:"script_path"`
}

type ScriptExec struct {
	ScriptPath string `json:"script_path"`
	LogPath    string `json:"log_path"`
}

type ScriptResult struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
	Out   string    `json:"out"`
	Err   string    `json:"err"`
}
